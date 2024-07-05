package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type mySecrets struct {
	configLocation string
	dbCon          string
	dbUser         string
	dbPassword     string
}

func (secrets *mySecrets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<body><h1>I am a GO application running inside Kubernetes.</h1> <h2>My properties are:</h2>")

	fmt.Fprintf(w, "<p>I read my secrets from %s</p>", secrets.configLocation)

	fmt.Fprintf(w, "<h2> Database connection details</h2>")
	fmt.Fprintf(w, "<ul><li>%s</li>", secrets.dbCon)
	fmt.Fprintf(w, "<li>%s</li>", secrets.dbUser)
	fmt.Fprintf(w, "<li>%s</li>", secrets.dbPassword)
	fmt.Fprintf(w, "</ul></body>")

}

func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	secrets := mySecrets{}
	secrets.readCurrentConfiguration()

	http.Handle("/", &secrets)

	// Kubernetes check if app is ok
	http.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "up")
	})

	// Kubernetes check if app can serve requests
	http.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "yes")
	})

	fmt.Printf("Simple web server is listening now at port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func (secrets *mySecrets) readCurrentConfiguration() {
	viper.SetDefault("db_con", "mysql.example.com:3306")
	viper.SetDefault("db_user", "demoUser")
	viper.SetDefault("db_password", "demoPassword")

	viper.SetConfigName("credentials")
	viper.SetConfigType("properties")

	//Development mode
	viper.AddConfigPath(".")

	//Proper configuration in non-development mode
	viper.AddConfigPath("/secrets/")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	//Reload configuration when file changes
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		secrets.reloadSettings()

	})

	secrets.reloadSettings()

	viper.WatchConfig()

}

func (secrets *mySecrets) reloadSettings() {

	secrets.configLocation = viper.ConfigFileUsed()
	fmt.Printf("Reading configuration from %s\n", secrets.configLocation)

	secrets.dbCon = unQuoteIfNeeded(viper.GetString("db_con"))
	secrets.dbUser = unQuoteIfNeeded(viper.GetString("db_user"))
	secrets.dbPassword = unQuoteIfNeeded(viper.GetString("db_password"))

	fmt.Printf("Connection string is %s\n", secrets.dbCon)
	fmt.Printf("Username is %s\n", secrets.dbUser)
	fmt.Printf("Password is %s\n", secrets.dbPassword)

}

func unQuoteIfNeeded(input string) string {
	result := ""
	if strings.HasPrefix(input, "\"") {
		result, _ = strconv.Unquote(input)
	}
	return result
}
