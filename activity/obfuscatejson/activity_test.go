package obfusactejson

import (
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	ref := activity.GetRef(&Activity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

var payload string

func TestEval(t *testing.T) {
	settings := &Settings{Operation: "setLastFour", Fields: []interface{}{"LoyaltyRewardsNumber", "BookingCreditCard"}}

	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)

	payload = `{"application/json":{"flightTrack":{"flightId":271143235,"carrier":{"fs":"EK","name":"Emirates","phoneNumber":"1-800-777-3999","active":true},"CustomerProfile":{"FirstName":"Arden","LastName":"Kaur","LoyaltyRewardsNumber":"EK2340983419","BookingCreditCard":"41462917261957261"},"flightNumber":"202","tailNumber":"N774AN","callsign":"EK202","departureAirport":{"fs":"JFK","iata":"JFK","icao":"KJFK","faa":"JFK","name":"John F. Kennedy International Airport","street1":"JFK Airport","street2":"","city":"New York","cityCode":"NYC","stateCode":"NY","postalCode":"11430","countryCode":"US","countryName":"United States","regionName":"North America","timeZoneRegionName":"America/New_York","weatherZone":"NYZ076","localTime":"2020-08-09T14:58:44.106","utcOffsetHours":-4,"latitude":40.642335,"longitude":-73.78817,"elevationFeet":13,"classification":1,"active":true},"arrivalAirport":{"fs":"EK","name":"Dubai International Airport","city":"Dubai","utcOffsetHours":1,"latitude":51.469603,"longitude":-0.453566,"elevationFeet":80,"classification":1,"active":true},"departureDate":{"dateLocal":"2020-08-08T18:10:00.000","dateUtc":"2020-08-08T22:10:00.000Z"},"equipment":"777","delayMinutes":1,"bearing":119.04182593265193,"heading":89.9998044218202,"positions":[{"lon":-0.4657000005245209,"lat":51.47380065917969,"speedMph":154,"altitudeFt":360,"source":"ADS-B","date":"2020-08-09T05:13:13.000Z"},{"lon":-0.46619999408721924,"lat":51.47380065917969,"speedMph":154,"altitudeFt":360,"source":"ADS-B","date":"2020-08-09T05:12:52.000Z"},{"lon":-0.46650001406669617,"lat":51.47380065917969,"speedMph":154,"altitudeFt":360,"source":"ADS-B","date":"2020-08-09T05:12:33.000Z"},{"lon":-0.4668999910354614,"lat":51.47380065917969,"speedMph":154,"altitudeFt":360,"source":"ADS-B","date":"2020-08-09T05:12:23.000Z"}]}}}`

	tc := test.NewActivityContext(act.Metadata())
	input := &Input{Payload: payload}
	err = tc.SetInputObject(input)
	assert.Nil(t, err)

	done, err := act.Eval(tc)
	assert.True(t, done)
	assert.Nil(t, err)

	val := tc.GetOutput("result")
	assert.NotNil(t, val)

	assert.Contains(t, val.(string), "*************7261", "It obfuscated the digits of BookingCreditCard")
}
