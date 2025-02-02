foo=$(get_context "foo")
echo "Got 'foo' context val of ${foo}"

set_context "bar" "baz"

bar=$(get_context "bar")
echo "Got 'bar' context val of ${bar}"

echo "${bar}" > outputs/baz.txt
