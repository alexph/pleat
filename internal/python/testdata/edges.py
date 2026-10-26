@instrument
@trace("edges")
def edges(items):
    """Keep the decorator and docstring with the definition.

    This blank line is inside the docstring.
    """
    ready = check(
        items,

        strict=True,
    )
    # This attached comment does not separate the assignment and condition.
    if ready:
        value = normalize(items)
        return value
    elif items:
        return items
    else:
        return None

    result = collect(
        items,

        ready,
    )

    return result
