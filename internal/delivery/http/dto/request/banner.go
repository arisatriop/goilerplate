package dtorequest

type BannerListRequest struct {
	StoreID string `query:"store_id"`
}

type CreateBannerRequest struct {
	// Files will be handled via multipart form, not JSON
	// This struct can be used for additional metadata if needed
}
