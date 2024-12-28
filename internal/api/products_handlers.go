package api

import (
	"context"
	"net/http"

	"github.com/erikgmatos/gobid/internal/jsonutils"
	"github.com/erikgmatos/gobid/internal/services"
	"github.com/erikgmatos/gobid/internal/usecase/product"
	"github.com/google/uuid"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductReq](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	userID, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{
			"error": "unexpected error try again later",
		})
		return
	}
	productId, err := api.ProductServices.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{
			"error": "failed to create product auction try again later",
		})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRomm := services.NewAuctionRoom(ctx, productId, api.BidsServices)

	go auctionRomm.Run()
	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[productId] = auctionRomm
	api.AuctionLobby.Unlock()

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message":    "auction ha started with success",
		"product_id": productId,
	})
}
