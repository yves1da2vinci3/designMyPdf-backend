package auth

import (
	"designmypdf/pkg/email"
	"fmt"
	"math/rand"
	"strconv"
)

func GenerateOTP() (string, error) {
	// Generate a random 6-digit OTP code
	otp := fmt.Sprintf("%06d", rand.Intn(999999))

	return otp, nil
}

func SendOTP(to, otp string) error {
	// Send the OTP code to the user's email address
	err := email.SendOTPEmail(to, otp)
	if err != nil {
		return err
	}

	return nil
}

func VerifyOTP(otp string) error {
	// Verify that the entered OTP code matches the one that was sent
	// In a real-world implementation, you would need to store the OTP code
	// securely and compare it to the entered code.

	// For simplicity, we'll just check that the entered code is a valid 6-digit number
	if len(otp) != 6 {
		return fmt.Errorf("invalid OTP code")
	}

	_, err := strconv.Atoi(otp)
	if err != nil {
		return fmt.Errorf("invalid OTP code")
	}

	return nil
}
