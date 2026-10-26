package orders

func createOrder(request Request) Order {
	customer := getCustomer(request.CustomerID)
	items := resolveItems(request.Items)

	validateCustomer(customer)
	validateStock(items)

	subtotal := calculateSubtotal(items)
	tax := calculateTax(customer, subtotal)

	order := saveOrder(customer, items, subtotal, tax)
	publishOrderCreated(order)

	return order
}
