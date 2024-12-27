package product

import (
	"context"
	"time"

	"github.com/erikgmatos/gobid/internal/validator"
	"github.com/google/uuid"
)

type CreateProductReq struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionDuration = 2 * time.Hour

func (req CreateProductReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "this field cannot be blank")

	eval.CheckField(validator.NotBlank(req.Description), "description", "this field cannot be blank")
	eval.CheckField(validator.MinChars(req.Description, 10) &&
		validator.MaxChars(req.Description, 255), "description", "this field must have a lenght between 10 and 255 characters")

	eval.CheckField(req.BasePrice > 0, "base_price", "this field must be greater than 0")

	eval.CheckField(time.Until(req.AuctionEnd) >= minAuctionDuration, "auction_end", "must be at least 2 hours duration")

	return eval
}
