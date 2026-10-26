function trustworthy() {
  first();
  second();
  third();
  fourth();
  fifth();
  sixth();
  seventh();
  eighth();
  return result;
}

function broken() {
  const value = ;
  first();
  second();
  third();
  return value;
}

function alsoTrustworthy() {
  const value = read();
  use(value);
  return value;
}
