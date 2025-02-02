foo = get_context("foo")
print(f"Got 'foo' context val of {foo}")

set_context("bar", "baz")

bar = get_context("bar")
print(f"Got 'bar' context val of {bar}")

with open('outputs/baz.txt', 'w') as out_file:
    out_file.write(bar)
