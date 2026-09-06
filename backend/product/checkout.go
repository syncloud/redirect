package product

type Checkout interface {
	Start(order *Order, description string) (string, string, error)
	Paid(providerReference string) (bool, int, string, error)
}

type Checkouts struct {
	paypal Checkout
	stripe Checkout
}

func NewCheckouts(paypal Checkout, stripe Checkout) *Checkouts {
	return &Checkouts{paypal: paypal, stripe: stripe}
}

func (c *Checkouts) Get(provider string) (Checkout, error) {
	switch provider {
	case "paypal":
		return c.paypal, nil
	case "stripe":
		return c.stripe, nil
	}
	return nil, ErrUnknownProvider
}
