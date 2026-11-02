package handler

// The bar routes are mounted for three audiences, with the same handlers and different credentials.
// swag gives one annotated function one set of security rules, and the methods in bar.go document
// /api/v1. These empty functions carry the annotations for the other two audiences.

// barInternalCreateDoc documents POST /internal/bars.
//
// @Summary      Create bar (internal)
// @ID           internalBarsCreate
// @Description  Same behaviour as the /api/v1 route, for other services in the cluster: no user and no permission check.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        X-Service-Name  header  string  false  "Calling service, recorded in logs. An unverified claim"
// @Param        request  body  dtorequest.BarCreateRequest  true  "Bar data"
// @Success      201  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Header       201  {string}  Location  "Path of the created bar"
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401  {object}  response.BaseResponse
// @Failure      409  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     InternalSecret
// @Router       /internal/bars [post]
//
//nolint:unused // read only by swag
func barInternalCreateDoc() {}

// barInternalUpdateDoc documents PUT /internal/bars/{id}.
//
// @Summary      Update bar (internal)
// @ID           internalBarsUpdate
// @Description  Same behaviour as the /api/v1 route, for other services in the cluster: no user and no permission check.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        X-Service-Name  header  string  false  "Calling service, recorded in logs. An unverified claim"
// @Param        id  path  string  true  "Bar ID"
// @Param        request  body  dtorequest.BarUpdateRequest  true  "Bar data"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     InternalSecret
// @Router       /internal/bars/{id} [put]
//
//nolint:unused // read only by swag
func barInternalUpdateDoc() {}

// barInternalDeleteDoc documents DELETE /internal/bars/{id}.
//
// @Summary      Delete bar (internal)
// @ID           internalBarsDelete
// @Description  Same behaviour as the /api/v1 route, for other services in the cluster: no user and no permission check.
// @Tags         internal
// @Produce      json
// @Param        X-Service-Name  header  string  false  "Calling service, recorded in logs. An unverified claim"
// @Param        id  path  string  true  "Bar ID"
// @Success      204  "No Content"
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     InternalSecret
// @Router       /internal/bars/{id} [delete]
//
//nolint:unused // read only by swag
func barInternalDeleteDoc() {}

// barInternalListDoc documents GET /internal/bars.
//
// @Summary      List bars (internal)
// @ID           internalBarsList
// @Description  Same behaviour as the /api/v1 route, for other services in the cluster: no user and no permission check.
// @Tags         internal
// @Produce      json
// @Param        X-Service-Name  header  string  false  "Calling service, recorded in logs. An unverified claim"
// @Param        keyword  query  string  false  "Search keyword"
// @Param        page     query  int     false  "Page number"  default(1)
// @Param        limit    query  int     false  "Page size, capped at 100"  default(10)
// @Success      200  {object}  response.BaseResponse{data=[]dtoresponse.BarResponse,meta=response.Meta}
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     InternalSecret
// @Router       /internal/bars [get]
//
//nolint:unused // read only by swag
func barInternalListDoc() {}

// barInternalGetDoc documents GET /internal/bars/{id}.
//
// @Summary      Get bar by ID (internal)
// @ID           internalBarsGet
// @Description  Same behaviour as the /api/v1 route, for other services in the cluster: no user and no permission check.
// @Tags         internal
// @Produce      json
// @Param        X-Service-Name  header  string  false  "Calling service, recorded in logs. An unverified claim"
// @Param        id  path  string  true  "Bar ID"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     InternalSecret
// @Router       /internal/bars/{id} [get]
//
//nolint:unused // read only by swag
func barInternalGetDoc() {}

// barPartnerCreateDoc documents POST /partner/v1/bars.
//
// @Summary      Create bar (partner)
// @ID           partnerBarsCreate
// @Description  Same behaviour as the /api/v1 route, for an external partner. Rate limited per API key.
// @Tags         partner
// @Accept       json
// @Produce      json
// @Param        request  body  dtorequest.BarCreateRequest  true  "Bar data"
// @Success      201  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Header       201  {string}  Location  "Path of the created bar"
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401  {object}  response.BaseResponse
// @Failure      409  {object}  response.BaseResponse
// @Failure      429  {object}  response.BaseResponse
// @Header       429  {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     ApiKeyAuth
// @Router       /partner/v1/bars [post]
//
//nolint:unused // read only by swag
func barPartnerCreateDoc() {}

// barPartnerUpdateDoc documents PUT /partner/v1/bars/{id}.
//
// @Summary      Update bar (partner)
// @ID           partnerBarsUpdate
// @Description  Same behaviour as the /api/v1 route, for an external partner. Rate limited per API key.
// @Tags         partner
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Bar ID"
// @Param        request  body  dtorequest.BarUpdateRequest  true  "Bar data"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      400  {object}  response.BaseResponse{errors=[]response.ValidationErrorDetail}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      429  {object}  response.BaseResponse
// @Header       429  {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     ApiKeyAuth
// @Router       /partner/v1/bars/{id} [put]
//
//nolint:unused // read only by swag
func barPartnerUpdateDoc() {}

// barPartnerDeleteDoc documents DELETE /partner/v1/bars/{id}.
//
// @Summary      Delete bar (partner)
// @ID           partnerBarsDelete
// @Description  Same behaviour as the /api/v1 route, for an external partner. Rate limited per API key.
// @Tags         partner
// @Produce      json
// @Param        id  path  string  true  "Bar ID"
// @Success      204  "No Content"
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      429  {object}  response.BaseResponse
// @Header       429  {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     ApiKeyAuth
// @Router       /partner/v1/bars/{id} [delete]
//
//nolint:unused // read only by swag
func barPartnerDeleteDoc() {}

// barPartnerListDoc documents GET /partner/v1/bars.
//
// @Summary      List bars (partner)
// @ID           partnerBarsList
// @Description  Same behaviour as the /api/v1 route, for an external partner. Rate limited per API key.
// @Tags         partner
// @Produce      json
// @Param        keyword  query  string  false  "Search keyword"
// @Param        page     query  int     false  "Page number"  default(1)
// @Param        limit    query  int     false  "Page size, capped at 100"  default(10)
// @Success      200  {object}  response.BaseResponse{data=[]dtoresponse.BarResponse,meta=response.Meta}
// @Failure      401  {object}  response.BaseResponse
// @Failure      429  {object}  response.BaseResponse
// @Header       429  {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     ApiKeyAuth
// @Router       /partner/v1/bars [get]
//
//nolint:unused // read only by swag
func barPartnerListDoc() {}

// barPartnerGetDoc documents GET /partner/v1/bars/{id}.
//
// @Summary      Get bar by ID (partner)
// @ID           partnerBarsGet
// @Description  Same behaviour as the /api/v1 route, for an external partner. Rate limited per API key.
// @Tags         partner
// @Produce      json
// @Param        id  path  string  true  "Bar ID"
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.BarResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      429  {object}  response.BaseResponse
// @Header       429  {integer}  Retry-After  "Seconds until the limit resets"
// @Failure      500  {object}  response.BaseResponse
// @Security     ApiKeyAuth
// @Router       /partner/v1/bars/{id} [get]
//
//nolint:unused // read only by swag
func barPartnerGetDoc() {}
