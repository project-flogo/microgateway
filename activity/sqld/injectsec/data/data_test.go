// Copyright 2018 The InjectSec Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package data

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

func TestRegex(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	generators := TrainingDataGenerator(rnd)
	for _, generator := range generators {
		if generator.Regex != nil {
			parts := NewParts()
			generator.Regex(parts)
			exp, err := parts.Regex()
			if err != nil {
				t.Fatal(err)
			}
			regex, err := regexp.Compile(exp)
			if err != nil {
				panic(err)
			}
			form := strings.ToLower(generator.Form)
			if generator.Case != "" {
				form = strings.ToLower(generator.Case)
			}
			if !regex.MatchString(form) {
				t.Fatal(exp, form)
			}
		}
	}
}

func TestSample(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	generators := TrainingDataGenerator(rnd)
	for _, generator := range generators {
		if generator.Regex != nil {
			parts := NewParts()
			generator.Regex(parts)
			exp, err := parts.Regex()
			if err != nil {
				t.Fatal(err)
			}
			regex, err := regexp.Compile(exp)
			if err != nil {
				panic(err)
			}
			for i := 0; i < 1024; i++ {
				sample, err := parts.Sample(rnd)
				if err != nil {
					t.Fatal(err)
				}
				sample = strings.ToLower(sample)
				if !regex.MatchString(sample) {
					t.Fatal(exp, generator.Form, sample)
				}
			}
		}
	}
}
