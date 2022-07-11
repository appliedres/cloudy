package cloudy

import (
	"crypto/rand"
	"strings"

	"math/big"
	mrand "math/rand"

	"github.com/google/uuid"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

func init() {

}

func GeneratePassword(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	var password strings.Builder

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
		// random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random.Int64()]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		// random := rand.Intn(len(numberSet))
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))

		password.WriteString(string(numberSet[random.Int64()]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		// random := rand.Intn(len(upperCharSet))
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))

		password.WriteString(string(upperCharSet[random.Int64()]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		// random := rand.Intn(len(allCharSet))
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))

		password.WriteString(string(allCharSet[random.Int64()]))
	}
	inRune := []rune(password.String())
	mrand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

// Generate an ID. The ID will follow the pattern {prefix}-{id} where the id
// is a randomly generated string of alphanumeric characters
func GenerateId(prefix string, num int) string {
	if num <= 0 {
		num = 15
	}
	cnt := num - len(prefix) - 1

	id, _ := gonanoid.Generate("abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", cnt)
	return prefix + "-" + id
}

func GenerateRandom(num int) string {
	id, _ := gonanoid.Generate("abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", num)
	return id
}

func HashId(prefix string, parts ...string) string {
	data := strings.Join(parts, "---")

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(data))
	suffix := id.String()

	if prefix != "" {
		return prefix + "-" + suffix
	}

	return suffix
	// sha256 := sha256.New()
	// for _, part := range parts {
	// 	sha256.Write([]byte(part))
	// }
	// sum := sha256.Sum(nil)
	// return string(sum)
}
