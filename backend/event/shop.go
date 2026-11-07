package event

const (
	ShopView        = "shop.view"
	ShopSignedOut   = "shop.signedout"
	ShopOption      = "shop.option"
	ShopSpec        = "shop.spec"
	ShopPhoto       = "shop.photo"
	ShopSignIn      = "shop.signin"
	ShopContinue    = "shop.continue"
	ShopAddress     = "shop.address"
	ShopPayView     = "shop.pay.view"
	ShopPayStripe   = "shop.pay.stripe"
	ShopPayPayPal   = "shop.pay.paypal"
	ShopPayPalBroke = "shop.paypal.failed"
	ShopError       = "shop.error"
	ShopOrdered     = "shop.ordered"
)

func Shop() []string {
	return []string{
		ShopView,
		ShopSignedOut,
		ShopOption,
		ShopSpec,
		ShopPhoto,
		ShopSignIn,
		ShopContinue,
		ShopAddress,
		ShopPayView,
		ShopPayStripe,
		ShopPayPayPal,
		ShopPayPalBroke,
		ShopError,
		ShopOrdered,
	}
}
