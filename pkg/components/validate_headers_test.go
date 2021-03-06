package components

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func Test_incomingMatchesAllowed(t *testing.T) {
	defaultAllowedHeaderAndValues := map[string][]string{"Ldap-Groups": {"sre", "devs"}}
	defaultDelimiterValues := map[string]string{"Ldap-Groups": ","}
	tests := []struct {
		name           string
		allowedHeader  map[string][]string
		incomingHeader map[string][]string
		delimiter      map[string]string
		wantErr        bool
	}{
		{
			name:           "multiple incoming header values with multiple allowed values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre", "devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "multiple incoming header values with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"devs"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre", "devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "multiple incoming header values separated by a comma with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"devs"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre,ops", "security,devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "single incoming header value separated by a comma with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"ops"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"security,ops"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "single incoming header value with multiple allowed values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "single incoming header value with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"devs"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "single incoming header with a single allowed value and multiple allowed headers",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"sre"}, "Client": {"mobile"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        false,
		},
		{
			name:           "single incorrect incoming header value",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"design"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        true,
		},
		{
			name:           "multiple incorrect incoming header values with multiple allowed values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"design", "finance"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        true,
		},
		{
			name:           "missing specific incoming header",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"meh": {"sre", "devs"}},
			delimiter:      defaultDelimiterValues,
			wantErr:        true,
		},
		{
			name:           "missing incoming header and values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"": {""}},
			delimiter:      defaultDelimiterValues,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := incomingMatchesAllowed(tt.allowedHeader, tt.incomingHeader, tt.delimiter)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_contains(t *testing.T) {
	defaultS := []string{"dog", "cat", "bird", "fish"}
	tests := []struct {
		name         string
		sliceToCheck []string
		target       string
		wantResult   bool
	}{
		{
			name:         "target is not present",
			sliceToCheck: defaultS,
			target:       "insect",
			wantResult:   false,
		},
		{
			name:         "target is empty",
			sliceToCheck: defaultS,
			target:       "",
			wantResult:   false,
		},
		{
			name:         "target is present",
			sliceToCheck: defaultS,
			target:       "fish",
			wantResult:   true,
		},
		{
			name:         "target is present in a single value slice",
			sliceToCheck: []string{"dog"},
			target:       "dog",
			wantResult:   true,
		},
		{
			name:         "target is present in a slice with values separated by a comma",
			sliceToCheck: []string{"cat,dog"},
			target:       "dog",
			wantResult:   true,
		},
		{
			name:         "target is not present in a slice with values separated by a comma",
			sliceToCheck: []string{"cat,dog"},
			target:       "fish",
			wantResult:   false,
		},
		{
			name:         "target is present in a slice with multiple values separated by a comma",
			sliceToCheck: []string{"cat,dog", "fish,chicken"},
			target:       "dog",
			wantResult:   true,
		},
		{
			name:         "target is not present in a slice with multiple values separated by a comma",
			sliceToCheck: []string{"cat,dog", "fish,chicken"},
			target:       "bird",
			wantResult:   false,
		},
		{
			name:         "target is present in a slice with multiple values separated by a comma and a space",
			sliceToCheck: []string{"cat, dog", "fish, chicken"},
			target:       "dog",
			wantResult:   true,
		},
		{
			name:         "target is not present in a slice with multiple values separated by a comma and a space",
			sliceToCheck: []string{"cat, dog", "fish, chicken"},
			target:       "bird",
			wantResult:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, contains(tt.sliceToCheck, tt.target, ","), tt.wantResult)
		})
	}
}

func Test_validateHeadersRoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		allowedHeaders map[string][]string
		testHeaders    http.Header
		split          map[string]string
		wantErr        bool
		wantResponse   int
	}{
		{
			name:           "valid header and values present",
			allowedHeaders: map[string][]string{"client": {"mobile"}},
			testHeaders: http.Header{
				"Client": {"mobile", "browser"},
			},
			split:        map[string]string{},
			wantErr:      false,
			wantResponse: http.StatusOK,
		},
		{
			name:           "valid header and values present with split delimiter",
			allowedHeaders: map[string][]string{"client": {"browser", "mobile"}},
			testHeaders: http.Header{
				"Client": {"browser,mobile"},
			},
			split: map[string]string{"client": ","},

			wantErr:      false,
			wantResponse: http.StatusOK,
		},
		{
			name:           "valid header and values present with split delimiter, multi length delimiter",
			allowedHeaders: map[string][]string{"client": {"browser", "mobile"}},
			testHeaders: http.Header{
				"Client": {"browser,.-mobile"},
			},
			split: map[string]string{"client": ",.-"},

			wantErr:      false,
			wantResponse: http.StatusOK,
		},
		{
			name:           "missing allowed header value",
			allowedHeaders: map[string][]string{"client": {"browser"}},
			testHeaders: http.Header{
				"Client": {"telnet"},
			},
			split: map[string]string{},

			wantErr:      false,
			wantResponse: http.StatusBadRequest,
		},
		{
			name:           "empty delimiter for split success",
			allowedHeaders: map[string][]string{"client": {"browser"}},
			testHeaders: http.Header{
				"Client": {"browser", "mobile"},
			},
			split: map[string]string{"client": ""},

			wantErr:      false,
			wantResponse: http.StatusOK,
		},
		{
			name:           "empty delimiter for split failure",
			allowedHeaders: map[string][]string{"client": {"browser"}},
			testHeaders: http.Header{
				"Client": {"browsermobile"},
			},
			split: map[string]string{"client": ""},

			wantErr:      false,
			wantResponse: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rt := NewMockRoundTripper(ctrl)
			c := &validateHeaderTransport{
				Wrapped: rt,
				Allowed: tt.allowedHeaders,
				Split:   tt.split,
			}
			r := &http.Request{Header: tt.testHeaders}
			rt.EXPECT().RoundTrip(gomock.Any()).Return(
				&http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				},
				nil,
			).AnyTimes()
			got, err := c.RoundTrip(r)
			require.Equal(t, err != nil, tt.wantErr)
			require.Equal(t, tt.wantResponse, got.StatusCode)
		})
	}
}
