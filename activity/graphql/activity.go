package graphql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	graphqlgo "github.com/graph-gophers/graphql-go"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/microgateway/activity/graphql/ratelimiter"
)

const (
	// GqlModeA GraphQL policy based on input query depth
	GqlModeA = "a"
	// GqlModeB GraphQL policy based on utilized server time
	GqlModeB = "b"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// New creates new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	if settings.Mode == "a" {
		return &Activity{mode: settings.Mode}, nil
	}

	// mode "b":
	// validate settings
	_, _, _, err = ratelimiter.ParseLimitString(settings.Limit)
	if err != nil {
		return nil, err
	}

	logger := ctx.Logger()
	logger.Debugf("Setting: %b", settings)

	rLimiters := make(map[string]*ratelimiter.Limiter, 1)
	t1s := make(map[string]time.Time)

	act := &Activity{
		mode:  settings.Mode,
		limit: settings.Limit,
		context: Context{
			rateLimiters: rLimiters,
			t1s:          t1s,
		},
	}

	return act, nil
}

// Activity is an GraphQLActivity
// inputs : {message}
// outputs: none
type Activity struct {
	mode    string
	limit   string
	context Context
}

// Context graphql context
type Context struct {
	rateLimiters map[string]*ratelimiter.Limiter
	t1s          map[string]time.Time
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - TBD
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	fmt.Println("Evaluate graphQL policies")
	// get inputs
	input := &Input{}
	ctx.GetInputObject(input)
	// fmt.Println("query: ", input.Query)
	// fmt.Println("schema file: ", input.SchemaFile)

	switch a.mode {
	case "a":
		// START of mode-a
		{
			// check schema file is provided or not
			if input.SchemaFile == "" {
				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := "Schema file is required"
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			// load schema
			schemaStr, err := ioutil.ReadFile(input.SchemaFile)
			if err != nil {
				fmt.Printf("Not able to read the schema file[%s] with the error - %s \n", input.SchemaFile, err)
				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := fmt.Sprintf("Not able to read the schema file[%s] with the error - %s \n", input.SchemaFile, err)
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}
			schema, err := graphqlgo.ParseSchema(string(schemaStr), nil)
			if err != nil {
				fmt.Println("Error while parsing graphql schema: ", err)
				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := fmt.Sprintf("Error while parsing graphql schema: %s", err)
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			// parse request
			var gqlQuery struct {
				Query         string                 `json:"query"`
				OperationName string                 `json:"operationName"`
				Variables     map[string]interface{} `json:"variables"`
			}
			err = json.Unmarshal([]byte(input.Query), &gqlQuery)
			if err != nil {
				fmt.Println("Error while parsing graphql query: ", err)
				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := fmt.Sprintf("Not a valid graphQL request. Details: %s", err)
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			// check query depth
			depth := calculateQueryDepth(gqlQuery.Query)
			if depth > input.MaxQueryDepth {
				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := fmt.Sprintf("graphQL request query depth[%v] is exceeded allowed maxQueryDepth[%v]", depth, input.MaxQueryDepth)
				fmt.Println(errMsg)
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			// validate request
			validationErrors := schema.Validate(gqlQuery.Query)
			if validationErrors != nil {
				fmt.Printf("Invalid graphql request: %s \n", validationErrors)

				// set error flag & error message in the output
				err = ctx.SetOutput("error", true)
				if err != nil {
					return false, err
				}
				errMsg := fmt.Sprintf("Not a valid graphQL request. Details: %s", validationErrors)
				err = ctx.SetOutput("errorMessage", errMsg)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			// set output
			err = ctx.SetOutput("valid", true)
			if err != nil {
				return false, err
			}
			validationMsg := fmt.Sprintf("Valid graphQL query. query = %s\n type = Query \n queryDepth = %v", input.Query, depth)
			err = ctx.SetOutput("validationMessage", validationMsg)
			if err != nil {
				return false, err
			}

			return true, nil
		}
		// END of mode-a

	case "b":
		// START of mode-b
		{
			var rateLimiterKey string
			rateLimiterKey = input.Token
			if rateLimiterKey == "" {
				rateLimiterKey = "GLOBAL_TOKEN"
			}

			switch input.Operation {
			case "startconsume":
				// get limiter, if not create one
				limiter, ok := a.context.rateLimiters[rateLimiterKey]
				if !ok {
					fmt.Printf("creating a ratelimiter with the limit[%s]\n", a.limit)
					limiter = ratelimiter.New(a.limit)
					a.context.rateLimiters[rateLimiterKey] = limiter
				}
				// check available limit
				if limiter.AvailableLimit() <= 0 {
					err = ctx.SetOutput("error", true)
					if err != nil {
						return false, err
					}
					errMsg := "Quota is not available"
					fmt.Println(errMsg)
					err = ctx.SetOutput("errorMessage", errMsg)
					if err != nil {
						return false, err
					}
				} else {
					err = ctx.SetOutput("error", false)
					if err != nil {
						return false, err
					}
				}

				// start time stamp
				a.context.t1s[rateLimiterKey] = time.Now()

			case "stopconsume":
				// elapsed time
				t1, ok := a.context.t1s[rateLimiterKey]
				if !ok {
					return false, nil
				}
				elapsed := time.Since(t1)
				elapsedInMs := 1000 * elapsed.Seconds()
				elapsedInMsRound := int(elapsedInMs)
				fmt.Printf("GraphQL call took %v ms\n", elapsedInMsRound)

				limiter, ok := a.context.rateLimiters[rateLimiterKey]
				if !ok {
					return false, nil
				}
				availableLimit, err := limiter.Consume(elapsedInMsRound)
				fmt.Printf("available limit = %v \n", availableLimit)
				if err != nil {
					err = ctx.SetOutput("error", true)
					if err != nil {
						return false, err
					}
					errMsg := "Consumed entire Quota"
					fmt.Println(errMsg)
					err = ctx.SetOutput("errorMessage", errMsg)
					if err != nil {
						return false, err
					}
				}

			default:
				return true, nil
			}
		}
		// START of mode-b
	default:
		return true, nil
	}

	return true, nil
}
