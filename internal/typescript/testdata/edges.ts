function smallReturn() {
  const value = read();
  return value;
}

function directIf() {
  const ready = read();
  if (ready) {
    use(ready);
  } else if (fallback()) {
    recover();
  } else {
    stop();
  }
}

function directLoop() {
  const items = read();
  for (const item of items) {
    use(item);
  }
}

function comments() {
  first();
  // A comment is not a blank separator.
  second();
  /* A multiline comment.

     Its interior blank line is still a comment. */
  third();
  fourth();
}

function multiline() {
  const value = transform<Result>(
    read(),

    options(),
  );
  return value;
}

function nested() {
  function inner() {
    first();
    second();
    third();
    fourth();
  }

  const callback = <T>(value: T) => {
    first();
    second();
    third();
    return value;
  };

  return callback;
}

function twoAssignments() {
  const first = read();
  const second = read();
  if (second) {
    use(first);
  }
}

function attachedComment() {
  prepare();
  // This stays attached to the following control.
  if (ready()) {
    run();
  }
}

function separatedComment() {
  prepare();

  // The separator precedes the attached comment.
  if (ready()) {
    run();
  }
}

function destructured() {
  const {ready: active, ...rest} = read();
  if (active && rest) {
    run();
  }
}

function clauses() {
  try {
    run();
  } catch (error) {
    recover(error);
  } finally {
    close();
  }
}
