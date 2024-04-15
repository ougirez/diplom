package controller

//func (c *Controller) SignupUser(ctx echo.Context) error {
//	request := &domain.SignupUserRequest{}
//	if err := ctx.Bind(request); err != nil {
//		return err
//	}
//
//	response, err := c.service.SignupUser(ctx.Request().Context(), request)
//	if err != nil {
//		return err
//	}
//
//	ctx.SetCookie(utils.CreateHttpOnlyCookie(constants.CookieKeyAuthToken, response.AuthToken, viper.GetInt64(constants.ViperJWTTTLKey)))
//
//	return ctx.JSON(http.StatusOK, response)
//}
//
//func (c *Controller) LoginUser(ctx echo.Context) error {
//	request := &domain.LoginUserRequest{}
//	if err := ctx.Bind(request); err != nil {
//		return err
//	}
//
//	response, err := c.service.LoginUser(ctx.Request().Context(), request)
//	if err != nil {
//		return err
//	}
//
//	ctx.SetCookie(utils.CreateHttpOnlyCookie(constants.CookieKeyAuthToken, response.AuthToken, viper.GetInt64(constants.ViperJWTTTLKey)))
//
//	return ctx.JSON(http.StatusOK, response)
//}
//
//func (c *Controller) LogoutUser(ctx echo.Context) error {
//	ctx.SetCookie(utils.CreateHttpOnlyCookie(constants.CookieKeyAuthToken, "", 0))
//	ctx.SetCookie(utils.CreateHttpOnlyCookie(constants.CookieKeySecretToken, "", 0))
//	return ctx.JSON(http.StatusOK, &domain.BasicResponse{})
//}
