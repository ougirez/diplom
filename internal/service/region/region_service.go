package region

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff/v4"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/pkg/store"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	store store.Store
}

func NewRegionItemService(regionItemDAO store.Store) *Service {
	return &Service{store: regionItemDAO}
}

func (s *Service) ParseAndSaveRegionItems(ctx context.Context, mainURL string) ([]*domain.RegionItem, error) {
	// Отправляем GET запрос
	resp, err := http.Get(mainURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get main page: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	// Проверяем статус ответа, он должен быть 200 OK
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Используем goquery для парсинга страницы
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}

	regionItems := make([]*domain.RegionItem, 0, 100)
	regionItemsMx := sync.Mutex{}
	eg, egCtx := errgroup.WithContext(ctx)
	doc.Find("div#block-mcxdm-mcxdm-system-main article table").EachWithBreak(func(i int, table *goquery.Selection) bool {
		districtName := table.Find("caption.fgbu-h2").Text()

		table.Find("tbody tr").EachWithBreak(func(i int, tr *goquery.Selection) bool {
			regionName := tr.Find("th").Text()
			regionHref, ok := tr.Find("td a").Attr("href")
			if !ok {
				err = fmt.Errorf("couldn't find href for region %s", regionName)
				return false
			}

			id := strings.Split(regionHref, "/")[len(strings.Split(regionHref, "/"))-1]

			eg.Go(func() error {
				regionItem, err := s.parseRegionItem(egCtx, fmt.Sprintf("%s/%s", mainURL, id))
				if err != nil {
					return fmt.Errorf("parseRegionItem, id-%s: %w", id, err)
				}

				regionItem.ID = id
				regionItem.RegionName = regionName
				regionItem.DistrictName = districtName

				err = s.store.Insert(ctx, regionItem)
				if err != nil {
					return fmt.Errorf("store.Insert, region_name-%s: %w", regionName, err)
				}

				log.Printf("parsed info for %s\n", regionName)

				regionItemsMx.Lock()
				defer regionItemsMx.Unlock()
				regionItems = append(regionItems, regionItem)
				return nil
			})

			return true
		})
		if err != nil {
			return false
		}

		return true
	})

	err = eg.Wait()
	if err != nil {
		return nil, fmt.Errorf("err in goroutine: %w", err)
	}

	return regionItems, nil
}

func (s *Service) parseRegionItem(ctx context.Context, regionURL string) (*domain.RegionItem, error) {
	// Отправляем GET запрос
	resp, err := http.Get(regionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get region doc: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	// Проверяем статус ответа, он должен быть 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Используем goquery для парсинга страницы
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}

	regionItem, err := parseRegionPage(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("parseRegionPage: %w", err)
	}

	return regionItem, nil
}

func parseRegionPage(ctx context.Context, doc *goquery.Document) (*domain.RegionItem, error) {
	// Extract data by year
	regionItem := new(domain.RegionItem)
	regionItem.GroupedCategories = make(map[string]*domain.GroupedCategory)

	var err error
	doc.Find("ol#fr-main li").EachWithBreak(func(_ int, li *goquery.Selection) bool {
		yearStr := li.Find("span.year").Text()
		if yearStr != "" {
			year, parseErr := strconv.Atoi(yearStr)
			if parseErr != nil {
				err = fmt.Errorf("failed to parse year: %w", parseErr)
				return false
			}

			fillErr := fillMainTableDataForYear(li, regionItem, year)
			if fillErr != nil {
				err = fmt.Errorf("fillMainTableDataForYear: %w", err)
				return false
			}
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	eg, egCtx := errgroup.WithContext(ctx)

	doc.Find("td.map-cols-l-b ul li").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		if i == 1 {
			selection.Find("a").EachWithBreak(func(i int, a *goquery.Selection) bool {
				href, ok := a.Attr("href")
				if !ok {
					// скипаем
					return true
				}

				eg.Go(func() error {
					yearStr := a.Text()
					year, err := strconv.Atoi(yearStr)
					if err != nil {
						return fmt.Errorf("failed to parse year: %w", err)
					}

					err = fillIrrigationIndicators(egCtx, regionItem, "https://inform-raduga.ru"+href, year)
					if err != nil {
						return fmt.Errorf("fillIrrigationIndicators: %w", err)
					}

					return nil
				})

				return true
			})
		} else if i == 3 {
			//selection.Find("a").Each(func(i int, a *goquery.Selection) {
			//	wg.Add(1)
			//
			//	href, ok := a.Attr("href")
			//	if !ok {
			//		return
			//	}
			//
			//	yearStr := a.Text()
			//	year, err := strconv.Atoi(yearStr)
			//	if err != nil {
			//		panic(err)
			//	}
			//
			//	go fillHistoryData(regionItem, &wg, "https://inform-raduga.ru"+href, year)
			//})
		}

		return true
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	return regionItem, nil
}

func fillMainTableDataForYear(li *goquery.Selection, regionItem *domain.RegionItem, year domain.Year) error {
	cols := []string{
		"Фактически используются в с/х производстве",
		"Фактически полито/осушено",
		"Залежные земли Всего",
		"Залежные земли Бесхозяйные",
		"Всего мелиорированных земель (с/х угодья)",
		"Всего мелиорированных земель (прочие земли)",
	}

	var err error
	// пробегаемся по строкам
	li.Find("tr").EachWithBreak(func(_ int, tr *goquery.Selection) bool {
		// The first th in the table is the category of this table
		groupCategoryName := tr.Find("th[scope=rowgroup]").Text()
		if groupCategoryName == "" || strings.Contains(groupCategoryName, "Итого") {
			// скипаем
			return true
		}

		groupCategoryName = strings.ReplaceAll(groupCategoryName, ", тыс. га", "")

		groupCategory := regionItem.GetGroupCategory(groupCategoryName)

		// пробегаемся по ячейкам в строке
		tr.Find("td:not(.t-outlined)").EachWithBreak(func(i int, td *goquery.Selection) bool {
			ul := td.Find("ul")
			if ul.Length() > 0 {
				total := decimal.Decimal{}

				// проходимся по всем подкатегориям из ячейки таблицы
				ul.Find("li").EachWithBreak(func(_ int, li *goquery.Selection) bool {
					liData := strings.Split(li.Text(), " — ")

					val, parseErr := strconv.ParseFloat(strings.ReplaceAll(liData[1], ",", "."), 64)
					if parseErr != nil {
						err = fmt.Errorf("failed to parse liData: %w", err)
						return false
					}

					categoryFirstName := cols[i]
					categoryLastName := liData[0]
					categoryName := fmt.Sprintf("%s — %s", categoryFirstName, categoryLastName)

					total = total.Add(decimal.NewFromFloat(val))

					category := groupCategory.GetCategory(categoryName)
					putErr := category.PutData(year, val, "тыс. га")
					if putErr != nil {
						err = fmt.Errorf("category.PutData, year-%d, val-%f: %w", year, val, err)
						return false
					}

					return true
				})
				if err != nil {
					return false
				}

				totalCategoryName := fmt.Sprintf("%s — всего", cols[i])
				totalCategory := groupCategory.GetCategory(totalCategoryName)
				totalVal := total.Round(3).InexactFloat64()
				putErr := totalCategory.PutData(year, totalVal, "тыс. га")
				if putErr != nil {
					err = fmt.Errorf("category.PutData, year-%d, val-%f: %w", year, totalVal, err)
					return false
				}
			} else {
				val, parseErr := strconv.ParseFloat(strings.ReplaceAll(td.Text(), ",", "."), 64)
				if parseErr != nil {
					err = fmt.Errorf("failed to parse td: %w", parseErr)
					return false
				}

				categoryName := cols[i]
				category := groupCategory.GetCategory(categoryName)
				putErr := category.PutData(year, val, "тыс. га")
				if putErr != nil {
					err = fmt.Errorf("category.PutData, year-%d, val-%f: %w", year, val, err)
					return false
				}
			}

			return true
		})

		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	return nil
}

func fillIrrigationIndicators(ctx context.Context, regionItem *domain.RegionItem, url string, year domain.Year) (err error) {
	var resp *http.Response
	err = backoff.Retry(
		func() error {
			var httpErr error

			resp, httpErr = http.Get(url)
			if httpErr != nil {
				return fmt.Errorf("http.Get: %w", httpErr)
			}
			// Проверяем статус ответа, он должен быть 200 OK
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status code error: %d %s; %s", resp.StatusCode, resp.Status)
			}

			return nil
		},
		backoff.WithContext(
			backoff.WithMaxRetries(backoff.NewConstantBackOff(10*time.Millisecond), 10),
			ctx,
		),
	)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			err = fmt.Errorf("failed to close reader: %w", closeErr)
		}
	}()

	// Используем goquery для парсинга страницы
	doc, parseErr := goquery.NewDocumentFromReader(resp.Body)
	if parseErr != nil {
		return fmt.Errorf("goquery.NewDocumentFromReader: %w", parseErr)
	}

	groupCategoryNames := make([]string, 0, 8)
	doc.Find("table.fgbu-col2 tbody tr td table tbody tr").EachWithBreak(func(i int, tr *goquery.Selection) bool {
		ths := tr.Find("th")

		if ths.Length() == 2 {
			// новая групповая категория
			groupCategoryName := ths.Eq(1).Find("strong").Text()
			groupCategoryNames = append(groupCategoryNames, groupCategoryName)
			return true
		}

		index, parseErr := strconv.Atoi(strings.Split(ths.Eq(0).Text(), ".")[0])
		if parseErr != nil {
			err = fmt.Errorf("failed to parse th: %w", parseErr)
			return false
		}
		groupCategory := regionItem.GetGroupCategory(groupCategoryNames[index-1])

		categoryName := strings.TrimSpace(ths.Eq(1).Text())

		valStr := strings.ReplaceAll(strings.TrimSpace(tr.Find("td").Text()), ",", ".")
		if strings.HasPrefix(categoryName, "...") {
			// скипаем
			return true
		}
		if valStr == "" {
			valStr = "0"
		}

		val, parseErr := strconv.ParseFloat(strings.ReplaceAll(valStr, " ", ""), 64)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse val: %w", parseErr)
			return false
		}

		category := groupCategory.GetCategory(categoryName)
		putErr := category.PutData(year, val, "тыс. га")
		if putErr != nil {
			err = fmt.Errorf("category.PutData, year-%d, val-%f: %w", year, val, err)
			return false
		}

		return true
	})

	return
}

//func fillHistoryData(regionItem *domain.RegionItem, wg *sync.WaitGroup, url string, year domain.Year) {
//	defer wg.Done()
//
//	resp, err := http.Get(url)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer resp.Body.Close()
//
//	// Проверяем статус ответа, он должен быть 200 OK
//	if resp.StatusCode != http.StatusOK {
//		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
//	}
//
//	// Используем goquery для парсинга страницы
//	doc, err := goquery.NewDocumentFromReader(resp.Body)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	groupCategory, ok := regionItem.GroupedCategories["Исторические показатели"]
//	if !ok {
//		groupCategory = &domain.GroupedCategory{Categories: make(map[string]*domain.Category)}
//		regionItem.GroupedCategories["Исторические показатели"] = groupCategory
//	}
//
//	doc.Find("table.fgbu-passport tbody tr").Each(func(i int, tr *goquery.Selection) {
//		categoryName := strings.TrimSpace(tr.Find("th").First().Text())
//
//		var unit string
//		if categoryName == "Годовой объем забираемой воды из различных водных объектов для орошения (водопотребление), млн. м3" {
//			unit = "млн. м3"
//			categoryName = strings.ReplaceAll(categoryName, ", млн. м3", "")
//		} else if i >= 23 {
//			unit = "тыс. га"
//			categoryName = strings.ReplaceAll(categoryName, ", тыс. га", "")
//		}
//
//		category, ok := groupCategory.Categories[categoryName]
//		if !ok {
//			category = &domain.Category{
//				YearData: make(map[domain.Year]float64),
//				Unit:     unit,
//			}
//			groupCategory.Categories[categoryName] = category
//		}
//
//		valStr := strings.ReplaceAll(strings.TrimSpace(tr.Find("td").First().Text()), ",", ".")
//		if valStr == "" {
//			valStr = "0"
//		}
//
//		val, err := strconv.ParseFloat(valStr, 64)
//		if err != nil {
//			return
//		}
//
//		if _, ok := category.YearData[year]; !ok {
//			category.YearData[year] = val
//		}
//	})
//}
