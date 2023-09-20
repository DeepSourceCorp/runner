package github

// func TestAPIProxyFactory_NewClient(t *testing.T) {
// 	testCases := []struct {
// 		name                   string
// 		appID                  string
// 		installationID         string
// 		apps                   map[string]*App
// 		expectError            bool
// 		expectedApp            *App
// 		expectedClient         *http.Client
// 		expectedInstallationID string
// 	}{
// 		{
// 			name:                   "valid app id",
// 			appID:                  "test-app-id",
// 			installationID:         "test-installation-id",
// 			apps:                   map[string]*App{"test-app-id": {ID: "test-app-id"}},
// 			expectError:            false,
// 			expectedApp:            &App{ID: "test-app-id"},
// 			expectedClient:         http.DefaultClient,
// 			expectedInstallationID: "test-installation-id",
// 		},
// 		{
// 			name:                   "invalid app id",
// 			appID:                  "test-app-id",
// 			installationID:         "test-installation-id",
// 			apps:                   map[string]*App{},
// 			expectError:            true,
// 			expectedApp:            nil,
// 			expectedClient:         nil,
// 			expectedInstallationID: "",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			factory := NewAPIProxyFactory(tc.apps, http.DefaultClient)
// 			proxy, err := factory.NewProxy(tc.appID, tc.installationID)
// 			if tc.expectError {
// 				if err == nil {
// 					t.Errorf("APIProxyFactory.NewClient() error = %v, want not nil", err)
// 				}
// 				return
// 			}

// 			if err != nil {
// 				t.Errorf("APIProxyFactory.NewClient() error = %v", err)
// 				return
// 			}

// 			if !reflect.DeepEqual(proxy.app, tc.expectedApp) {
// 				t.Errorf("APIProxyFactory.NewClient() app = %v, want %v", proxy.app, tc.expectedApp)
// 			}

// 			if proxy.installationID != tc.expectedInstallationID {
// 				t.Errorf("APIProxyFactory.NewClient() installationID = %v, want %v", proxy.installationID, tc.expectedInstallationID)
// 			}

// 			if proxy.client == nil {
// 				t.Errorf("APIProxyFactory.NewClient() client = %v, want not nil", proxy.client)
// 			}
// 		})
// 	}
// }

// func TestAPIProxy_GenerateJWT(t *testing.T) {
// 	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
// 	app := &App{ID: "test-app-id", PrivateKey: privateKey}
// 	proxy := &APIProxy{app: app}
// 	token, err := proxy.GenerateJWT()
// 	if err != nil {
// 		t.Errorf("APIProxy.GenerateJWT() error = %v", err)
// 		return
// 	}

// 	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
// 		publicKey := privateKey.PublicKey
// 		return &publicKey, nil
// 	})
// 	if err != nil {
// 		t.Errorf("APIProxy.GenerateJWT() error = %v", err)
// 		return
// 	}

// 	if parsedToken.Valid != true {
// 		t.Errorf("APIProxy.GenerateJWT() token is invalid")
// 		return
// 	}
// }

// func TestAPIProxy_GenerateAccessToken(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		responseStatus int
// 		responseBody   string
// 		expectedToken  string
// 		expectError    bool
// 	}{
// 		{
// 			name:           "valid response from Github",
// 			responseStatus: http.StatusCreated,
// 			responseBody:   `{"token": "test-token"}`,
// 			expectedToken:  "test-token",
// 			expectError:    false,
// 		},
// 		{
// 			name:           "invalid response from Github",
// 			responseStatus: http.StatusCreated,
// 			responseBody:   `{"invalid": "response"}`,
// 			expectedToken:  "",
// 			expectError:    true,
// 		},
// 		{
// 			name:           "invalid response status from Github",
// 			responseStatus: http.StatusInternalServerError,
// 			responseBody:   `{"token": "test-token"}`,
// 			expectedToken:  "",
// 			expectError:    true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(tc.responseStatus)
// 				_, _ = w.Write([]byte(tc.responseBody))
// 			}))
// 			defer server.Close()

// 			su, _ := url.Parse(server.URL)

// 			proxy := &APIProxy{
// 				client:         http.DefaultClient,
// 				installationID: "test-installation-id",
// 				app:            &App{APIHost: *su},
// 			}
// 			token, err := proxy.GenerateAccessToken("request-token")
// 			if tc.expectError {
// 				if err == nil {
// 					t.Errorf("APIProxy.GenerateAccessToken() error = %v, want not nil", err)
// 				}
// 				return
// 			}

// 			if err != nil {
// 				t.Errorf("APIProxy.GenerateAccessToken() error = %v", err)
// 				return
// 			}

// 			if token != tc.expectedToken {
// 				t.Errorf("APIProxy.GenerateAccessToken() token = %v, want %v", token, tc.expectedToken)
// 			}
// 		})
// 	}
// }

// func TestAPIProxy_ProxyURL(t *testing.T) {
// 	su, _ := url.Parse("https://example.com")

// 	// Create a mock APIProxy instance
// 	apiProxy := &APIProxy{
// 		app: &App{
// 			ID:      "your-app-id",
// 			APIHost: *su,
// 		},
// 	}

// 	// Define test cases
// 	tests := []struct {
// 		path             string
// 		expectedProxyURL string
// 	}{
// 		{
// 			path:             "/apps/your-app-id/api/endpoint",
// 			expectedProxyURL: "https://example.com/endpoint",
// 		},
// 		{
// 			path:             "/apps/your-app-id/api/another/endpoint",
// 			expectedProxyURL: "https://example.com/another/endpoint",
// 		},
// 		{
// 			path:             "/apps/your-app-id/api/",
// 			expectedProxyURL: "https://example.com",
// 		},
// 	}

// 	// Run test cases
// 	for _, test := range tests {
// 		proxyURL := apiProxy.ProxyURL(test.path)

// 		if proxyURL != test.expectedProxyURL {
// 			t.Errorf("ProxyURL(%s) = %s, expected %s", test.path, proxyURL, test.expectedProxyURL)
// 		}
// 	}
// }

// func TestAPIProxy_InstallationURL(t *testing.T) {
// 	baseURL, _ := url.Parse("https://example.com")
// 	app := &App{
// 		BaseHost: *baseURL,
// 		AppSlug:  "your-app-slug",
// 	}
// 	want := "https://example.com/apps/your-app-slug/installations/new"
// 	apiProxy := &APIProxy{
// 		app: app,
// 	}
// 	got := apiProxy.InstallationURL()
// 	if got != want {
// 		t.Errorf("InstallationURL() returned URL: %s, expected: %s", got, want)
// 	}
// }
