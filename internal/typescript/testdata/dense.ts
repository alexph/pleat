function createOrder(request: Request) {
  const customer = getCustomer(request.customerId);
  const items = resolveItems(request.items);
  validateCustomer(customer);
  validateStock(items);
  const subtotal = calculateSubtotal(items);
  const tax = calculateTax(customer, subtotal);
  const order = saveOrder(customer, items, subtotal, tax);
  publishOrderCreated(order);
  return order;
}
