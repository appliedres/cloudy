package cloudy

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"math/big"
	mrand "math/rand"

	"github.com/google/uuid"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	lowerCharSet   = "abcdefghijklmnopqrstuvwxyz"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*%"
	numberSet      = "0123456789"
)

type PasswordOptions struct {
	Length int

	MinUpperCase   int
	MinNum         int
	MinSpecialChar int

	HasUpperCase   bool
	HasNum         bool
	HasSpecialChar bool
}

func (po PasswordOptions) GetCharSet() string {
	charSet := lowerCharSet

	if po.HasUpperCase {
		charSet += upperCharSet
	}

	if po.HasNum {
		charSet += numberSet
	}

	if po.HasSpecialChar {
		charSet += specialCharSet
	}

	return charSet
}

func init() {

}

func GeneratePasswordNoSpecial(passwordLength, minNum, minUpperCase int) string {
	return GeneratePasswordFromOptions(PasswordOptions{
		Length:         passwordLength,
		MinUpperCase:   minUpperCase,
		MinNum:         minNum,
		HasUpperCase:   true,
		HasNum:         true,
		HasSpecialChar: false,
	})
}

func IsValidPassword(password string) bool {
	return IsValidPasswordWithOptions(password, PasswordOptions{
		HasUpperCase:   true,
		HasNum:         true,
		HasSpecialChar: true,
	})
}

func IsValidPasswordNoSpecial(password string) bool {
	return IsValidPasswordWithOptions(password, PasswordOptions{
		HasUpperCase:   true,
		HasNum:         true,
		HasSpecialChar: false,
	})
}

func IsValidPasswordWithOptions(password string, options PasswordOptions) bool {
	if len(password) == 0 {
		return false
	}

	var (
		lower   = regexp.MustCompile(fmt.Sprintf("[%s]{1}", lowerCharSet))
		upper   = regexp.MustCompile(fmt.Sprintf("[%s]{1}", upperCharSet))
		number  = regexp.MustCompile(fmt.Sprintf("[%s]{1}", numberSet))
		special = regexp.MustCompile(fmt.Sprintf("[%s]{1}", specialCharSet))
	)

	var (
		lowerFound   = len(lower.FindAllString(password, -1))
		upperFound   = len(upper.FindAllString(password, -1))
		numberFound  = len(number.FindAllString(password, -1))
		specialFound = len(special.FindAllString(password, -1))
		invalidFound = len(password) - lowerFound - upperFound - numberFound - specialFound
	)

	var (
		foundLower   = lowerFound >= 0 // Dont actually care about lower case
		foundUpper   = upperFound > 0
		foundNumber  = numberFound > 0
		foundSpecial = specialFound > 0
		foundInvalid = invalidFound > 0
	)

	if !options.HasSpecialChar {
		return foundUpper && foundLower && foundNumber && !foundSpecial && !foundInvalid
	}

	return foundUpper && foundLower && foundNumber && foundSpecial && !foundInvalid
}

func GeneratePassword(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	return GeneratePasswordFromOptions(PasswordOptions{
		Length: passwordLength,

		MinUpperCase:   minUpperCase,
		MinNum:         minNum,
		MinSpecialChar: minSpecialChar,
		HasUpperCase:   true,
		HasNum:         true,
		HasSpecialChar: true,
	})
}

func GeneratePasswordFromOptions(po PasswordOptions) string {
	var password strings.Builder

	if po.HasSpecialChar {
		//Set special character
		for i := 0; i < po.MinSpecialChar; i++ {
			random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
			// random := rand.Intn(len(specialCharSet))
			password.WriteString(string(specialCharSet[random.Int64()]))
		}
	}

	if po.HasNum {
		//Set numeric
		for i := 0; i < po.MinNum; i++ {
			// random := rand.Intn(len(numberSet))
			random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))

			password.WriteString(string(numberSet[random.Int64()]))
		}
	}

	if po.HasUpperCase {
		//Set uppercase
		for i := 0; i < po.MinUpperCase; i++ {
			// random := rand.Intn(len(upperCharSet))
			random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))

			password.WriteString(string(upperCharSet[random.Int64()]))
		}
	}

	charSet := po.GetCharSet()
	remainingLength := po.Length - po.MinSpecialChar - po.MinNum - po.MinUpperCase
	for i := 0; i < remainingLength; i++ {
		// random := rand.Intn(len(allCharSet))
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))

		password.WriteString(string(charSet[random.Int64()]))
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

// Generate an ID. The ID will follow the pattern {prefix}-{id} where the id
// is a randomly generated string of alphanumeric characters
func GenerateIdLower(prefix string, num int) string {
	if num <= 0 {
		num = 15
	}
	cnt := num - len(prefix) - 1

	id, _ := gonanoid.Generate("abcdefghijklmnopqrstuvxyz1234567890", cnt)
	return prefix + "-" + id
}

func GenerateRandom(num int) string {
	id, _ := gonanoid.Generate("abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", num)
	return id
}

func GenerateOtp(num int) string {
	id, _ := gonanoid.Generate("1234567890", num)
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

// returns a new VM ID prefixed with "uvm-".
func GenerateUVMID() string {
	id, _ := GenerateVMIDFromPrefix("uvm")
	return id
}

// returns a new VM ID prefixed with "shvm-".
func GenerateSHVMID() string {
	id, _ := GenerateVMIDFromPrefix("shvm")
	return id
}

// Using the prefix, generates a new VM ID is 15 or less characters.
// ID is a custom base36 encoded timestamp
// if prefix = 'uvm', returns 'uvm-0123456789'
func GenerateVMIDFromPrefix(prefix string) (string, error) {
	const (
		maxLen = 15
		sep    = "-"
	)
	idPart := GenerateTimestampID(time.Now())

	total := len(prefix) + len(sep) + len(idPart)
	if total > maxLen {
		return "", fmt.Errorf(
			"prefix %q too long: would produce ID length %d (max %d)",
			prefix, total, maxLen,
		)
	}

	return prefix + sep + idPart, nil
}

// GenerateID returns a 10‑char string: 
// 	9 for the millisecond timestamp in Base36 format,
// 	1 for a monotonic counter in the range 0–35. (to ensure uniqueness)
const timestampGenCounterMod = 36 // 0‑z  ⇒ one base‑36 digit
var timestampGenCounter uint32  // global timestampGenCounter (per‑process)
func GenerateTimestampID(t time.Time) string {
	// 1. Timestamp part (exactly 9 chars, zero‑padded)
	tsPart := strconv.FormatInt(t.UnixMilli(), 36)
	tsPart = strings.ToLower(fmt.Sprintf("%09s", tsPart))

	// 2. Counter part (exactly 1 char)
	n := atomic.AddUint32(&timestampGenCounter, 1) % timestampGenCounterMod
	ctPart := strconv.FormatUint(uint64(n), 36) // already lower

	return tsPart + ctPart
}

// GenerateID returns a 10‑char string: 
// 	9 for the millisecond timestamp in Base36 format,
// 	1 for a monotonic counter in the range 0–35. (to ensure uniqueness)
func GenerateTimestampIDNow() string {
	return GenerateTimestampID(time.Now())
}

// DecodeTime extracts the millisecond timestamp and converts it back to time.Time.
func DecodeTimestampID(id string) (time.Time, error) {
	if len(id) != 10 {
		return time.Time{}, fmt.Errorf("id must be 10 characters")
	}
	ms, err := strconv.ParseInt(id[:9], 36, 64) // first 9 chars
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(ms), nil
}