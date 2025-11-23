package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000
	return strconv.Itoa(otp)
}

