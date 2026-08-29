package product

import "fmt"

const Currency = "GBP"

func Money(pence int) string {
	return fmt.Sprintf("%d.%02d", pence/100, pence%100)
}
