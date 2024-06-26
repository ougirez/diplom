package service

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff/v4"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/domain/dto"
	"github.com/ougirez/diplom/internal/pkg/logger"
	"github.com/ougirez/diplom/internal/pkg/store"
	"github.com/samber/lo"
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

func NewProvidersService(store store.Store) *Service {
	return &Service{store: store}
}

func (s *Service) ParseAndSaveProviderItems(
	ctx context.Context,
	mainURL string,
) ([]*domain.Provider, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://inform-raduga.ru/fgbu", nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", " ")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get main page: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, resp.Body)
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}

	providerDtos := make([]*domain.Provider, 0, 100)
	providerDtosMx := sync.Mutex{}
	eg, egCtx := errgroup.WithContext(ctx)
	doc.Find("div#block-mcxdm-mcxdm-system-main article table").
		EachWithBreak(func(i int, table *goquery.Selection) bool {
			districtName := table.Find("caption.fgbu-h2").Text()

			table.Find("tbody tr").EachWithBreak(func(i int, tr *goquery.Selection) bool {
				eg.Go(func() error {
					regionName := tr.Find("th").Text()
					providerInfo := tr.Find("td a")

					providerHref, ok := providerInfo.Attr("href")
					if !ok {
						return fmt.Errorf("couldn't find href for providers %s", regionName)
					}

					id := strings.Split(providerHref, "/")[len(strings.Split(providerHref, "/"))-1]
					providerName := providerInfo.Text()
					providerDto, err := s.parseProviderItem(egCtx, fmt.Sprintf("%s/%s", mainURL, id))
					if err != nil {
						return fmt.Errorf("parseProviderItem, id-%s: %w", id, err)
					}

					idInt, err := strconv.Atoi(id)
					if err != nil {
						return fmt.Errorf("failed to parse id: %w", err)
					}

					providerDto.ProviderID = int64(idInt)
					providerDto.RegionName = regionName
					providerDto.DistrictName = districtName
					providerDto.ProviderName = providerName

					regionItem, err := s.store.Insert(egCtx, providerDto)
					if err != nil {
						log.Println(ctx, "store.Insert, region_name-%s: %w", regionName, err)
						return fmt.Errorf("store.Insert, region_name-%s: %w", regionName, err)
					}

					logger.Warnf(ctx, "parsed info for %s", regionName)

					providerDtosMx.Lock()
					defer providerDtosMx.Unlock()
					providerDtos = append(providerDtos, regionItem)
					return nil
				})

				return true
			})
			return true
		})

	err = eg.Wait()
	if err != nil {
		return nil, fmt.Errorf("err in goroutine: %w", err)
	}

	return providerDtos, nil
}

func (s *Service) parseProviderItem(ctx context.Context, providerURL string) (*dto.ProviderDto, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://inform-raduga.ru/fgbu", nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", " ")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers doc: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}

	regionItem, err := parseProviderPage(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("parseProviderPage: %w", err)
	}

	return regionItem, nil
}

func parseProviderPage(ctx context.Context, doc *goquery.Document) (*dto.ProviderDto, error) {
	providerDto := new(dto.ProviderDto)
	providerDto.GroupedCategories = make(map[string]*dto.GroupedCategory)

	var err error
	doc.Find("ol#fr-main li").EachWithBreak(func(_ int, li *goquery.Selection) bool {
		yearStr := li.Find("span.year").Text()
		if yearStr != "" {
			year, parseErr := strconv.Atoi(yearStr)
			if parseErr != nil {
				err = fmt.Errorf("failed to parse year: %w", parseErr)
				return false
			}

			fillErr := fillMainTableDataForYear(li, providerDto, year)
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
						logger.Errorf(ctx, "Atoi: %s", err.Error())
						return fmt.Errorf("failed to parse year: %w", err)
					}

					err = fillIrrigationIndicators(egCtx, providerDto, "https://inform-raduga.ru"+href, year)
					if err != nil {
						logger.Errorf(ctx, "fillIrrigationIndicators: %s", err.Error())
						return fmt.Errorf("fillIrrigationIndicators: %w", err)
					}

					return nil
				})

				return true
			})
		}

		return true
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	return providerDto, nil
}

func fillMainTableDataForYear(li *goquery.Selection, providerDto *dto.ProviderDto, year domain.Year) error {
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
		groupCategoryName := tr.Find("th[scope=rowgroup]").Text()
		if groupCategoryName == "" || strings.Contains(groupCategoryName, "Итого") {
			// скипаем
			return true
		}

		groupCategoryName = strings.ReplaceAll(groupCategoryName, ", тыс. га", "")

		groupCategory := providerDto.GetGroupCategory(groupCategoryName)

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

func fillIrrigationIndicators(ctx context.Context, providerDto *dto.ProviderDto, url string, year domain.Year) (err error) {
	var resp *http.Response
	err = backoff.Retry(
		func() error {
			var httpErr error

			resp, httpErr = http.Get(url)
			if httpErr != nil {
				return fmt.Errorf("http.Get: %w", httpErr)
			}
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
		groupCategory := providerDto.GetGroupCategory(groupCategoryNames[index-1])

		categoryName := strings.TrimSpace(ths.Eq(1).Text())
		unit := strings.ReplaceAll(strings.TrimSpace(ths.Eq(2).Text()), "\u00a0", " ")

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
		putErr := category.PutData(year, val, unit)
		if putErr != nil {
			err = fmt.Errorf("category.PutData, year-%d, val-%f: %w", year, val, err)
			return false
		}

		return true
	})

	return
}

func (s *Service) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	regionItems, err := s.store.ListRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("store.ListRegions: %w", err)
	}

	return regionItems, nil
}

func (s *Service) ListCategoriesByRegionID(ctx context.Context, opts store.ListCategoriesByRegionIDOpts) (map[string]map[string][]*domain.Category, error) {
	extendedCategories, err := s.store.ListCategoriesByRegionID(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("store.ListCategoriesByRegionID: %w", err)
	}

	res := make(map[string]map[string][]*domain.Category)
	for _, c := range extendedCategories {
		if _, ok := res[c.ProviderName]; !ok {
			res[c.ProviderName] = make(map[string][]*domain.Category)
		}
		res[c.ProviderName][c.GroupName] = append(res[c.ProviderName][c.GroupName], &c.Category)
	}

	return res, nil
}

type GetCategoryDataByRegionsResponse struct {
	CategoryName      string                     `json:"category_name"`
	GroupCategoryName string                     `json:"group_category_name"`
	Unit              string                     `json:"unit"`
	RegionsData       map[string]domain.YearData `json:"regions_data,omitempty"`
	MinYear           domain.Year                `json:"min_year,omitempty"`
	MaxYear           domain.Year                `json:"max_year,omitempty"`
}

func (s *Service) GetCategoryDataByRegions(ctx context.Context, opts store.GetCategoryDataByRegionsOpts) (*GetCategoryDataByRegionsResponse, error) {
	regionsData, err := s.store.GetCategoryDataByRegions(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("store.GetCategoryDataByRegions: %w", err)
	}

	resp := &GetCategoryDataByRegionsResponse{
		RegionsData:       make(map[string]domain.YearData),
		CategoryName:      opts.CategoryName,
		GroupCategoryName: opts.GroupCategoryName,
	}
	for _, r := range regionsData {
		resp.RegionsData[r.RegionName] = r.YearData

		keys := lo.Keys(r.YearData)
		minY := lo.Min(keys)
		maxY := lo.Max(keys)

		if resp.MinYear == 0 || minY < resp.MinYear {
			resp.MinYear = minY
		}
		if resp.MaxYear == 0 || maxY > resp.MaxYear {
			resp.MaxYear = maxY
		}
		resp.Unit = r.Unit
	}

	return resp, nil
}
