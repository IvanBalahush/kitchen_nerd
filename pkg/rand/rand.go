// RandStr helper function for random string generation
func RandStr(strSize int, randType Rand) (string, error) {
var dictionary string

switch randType {
case TypeAlphaNum:
dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
case TypeAlpha:
dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
case TypeNumber:
dictionary = "0123456789"
case TypeAllSymbols:
dictionary = "!\"#$%&\\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
case TypeToken:
dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789--------"
default:
return "", errs.New("invalid randType")
}

var bytes = make([]byte, strSize)
_, err := rand.Read(bytes)
if err != nil {
return "", errs.Wrap(err)
}

for k, v := range bytes {
bytes[k] = dictionary[v%byte(len(dictionary))]
}

return string(bytes), nil
}
