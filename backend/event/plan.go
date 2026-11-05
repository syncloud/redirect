package event

const (
	PlanView         = "plan.view"
	PlanSubscribed   = "plan.subscribed"
	PlanTier         = "plan.tier"
	PlanPeriod       = "plan.period"
	PlanStripe       = "plan.stripe"
	PlanPayPal       = "plan.paypal"
	PlanPayPalBroke  = "plan.paypal.failed"
	PlanCrypto       = "plan.crypto"
	PlanCryptoSent   = "plan.crypto.sent"
	PlanSwitchAnnual = "plan.switch.annual"
	PlanCancel       = "plan.cancel"
	PlanError        = "plan.error"
)

func Plan() []string {
	return []string{
		PlanView,
		PlanSubscribed,
		PlanTier,
		PlanPeriod,
		PlanStripe,
		PlanPayPal,
		PlanPayPalBroke,
		PlanCrypto,
		PlanCryptoSent,
		PlanSwitchAnnual,
		PlanCancel,
		PlanError,
	}
}
