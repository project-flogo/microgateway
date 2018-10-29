package jwt

import (
	//"errors"
	"fmt"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/logger"
)

const (
	ivServiceName = "serviceName"
	ivToken 	= "token"
	ivKey        	= "key"
	ivSigningMethod = "signingMethod"
	ivIssuer      	= "iss"
	ivSubject       = "sub"
	ivAudience      = "aud"

	ovValid   	= "valid"
	ovToken 	= "token"
	ovValidationMsg = "validationMessage"
	ovError    	= "error"
	ovErrorMsg 	= "errorMessage"
)

var (
	activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
	log = logger.GetLogger("activity-jwt")
)

func init() {
	activity.Register(&Activity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	log.Debugf("Setting: %b", settings)

	act := &Activity{}
	return act, nil
}

type Activity struct {}

// Metadata return the metadata for the activity
func (a *Activity) Metadata() *activity.Metadata {
	return activityMetadata
}

// Eval executes the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	fmt.Println("Inside eval")
	fmt.Println("context is :",ctx)
	input := Input{}
	fmt.Println("input is :",input)
	err = ctx.GetInputObject(&input)
	if err != nil {
		fmt.Println("inside error")
		return false, err
	}
	fmt.Println("here")
	token, err := jwt.Parse(input.Token, func(token *jwt.Token) (interface{}, error) {
		// Make sure signing alg matches what we expect
		switch strings.ToLower(input.SigningMethod) {
		case "hmac":
			fmt.Println("in hmac")
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "ecdsa":
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "rsa":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "rsapss":
			if _, ok := token.Method.(*jwt.SigningMethodRSAPSS); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "":
		// Just continue
		default:
			return nil, fmt.Errorf("Unknown signing method expected: %v", input.SigningMethod)
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if input.Issuer != "" && !claims.VerifyIssuer(input.Issuer, true) {
				fmt.Println("In Issuer")
				return nil, jwt.NewValidationError("iss claims do not match", jwt.ValidationErrorIssuer)
			}
			if input.Audience != "" && !claims.VerifyAudience(input.Audience, true) {
				fmt.Println("In Audience")
				return nil, jwt.NewValidationError("aud claims do not match", jwt.ValidationErrorAudience)
			}
			subClaim, sok := claims["sub"].(string)
			if input.Subject != "" && (!sok || strings.Compare(input.Subject, subClaim) != 0) {
				fmt.Println("In subject")
				return nil, jwt.NewValidationError("sub claims do not match", jwt.ValidationErrorClaimsInvalid)
			}
		} else {
			fmt.Println("in claims error")
			return nil, jwt.NewValidationError("unable to parse claims", jwt.ValidationErrorClaimsInvalid)
		}

		return []byte(ctx.GetInput(ivKey).(string)), nil
	})
	output := Output{}
	if token != nil && token.Valid {
		fmt.Println("valid token")
		output.Valid = true
		output.Token = ParsedToken{Signature: token.Signature, SigningMethod: token.Method.Alg(), Header: token.Header}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			output.Token.Claims = claims
		}
		return true,err
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		fmt.Println("invalid token")
		output.Valid = false
		output.ValidationMessage = ve.Error()
	} else {
		fmt.Println("errror in token")
		output.Valid = false
		output.Error = true
		output.ValidationMessage = err.Error()
		output.ErrorMessage = err.Error()
	}
	err = ctx.SetOutputObject(&output)
	if err != nil {
		return false, err
	}
	return true,nil
}
