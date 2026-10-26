def create_order(request):
    customer = get_customer(request.customer_id)
    items = resolve_items(request.items)

    validate_customer(customer)
    validate_stock(items)

    subtotal = calculate_subtotal(items)
    tax = calculate_tax(customer, subtotal)

    order = save_order(customer, items, subtotal, tax)
    publish_order_created(order)

    return order
