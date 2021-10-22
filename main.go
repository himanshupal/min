package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/himanshupal/min/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ctx    = context.Background()
	client *mongo.Client
)

const (
	indexPage    string = "./views/index.gohtml"
	linksPage    string = "./views/links.gohtml"
	notFoundPage string = "./views/not_found.gohtml"

	databaseName    string = "min_himanshupal_xyz"
	usersCollection string = "users"
	urlsCollection  string = "urls"
)

func getCollection(name string, w http.ResponseWriter) *mongo.Collection {
	var err error

	dbURI := os.Getenv("MONGO_URI")
	if dbURI == "" {
		dbURI = "mongodb://localhost"
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		fmt.Fprintln(w, err)
	}

	return client.Database(databaseName).Collection(name)
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.NoCache)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.StripSlashes)

	// Homepage
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		template.Must(template.ParseFiles(indexPage)).Execute(w, &bson.M{
			"Site": r.Host,
		})
	})

	// Create new account or redirect to user page
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()

		user.Username = r.PostFormValue("username")
		user.Password = r.PostFormValue("password")

		existingUser := r.PostFormValue("existingUser")

		if existingUser != "" {
			if err := getCollection(usersCollection, w).FindOne(ctx, &bson.M{"username": existingUser}).Err(); err == mongo.ErrNoDocuments {
				if err := client.Disconnect(ctx); err != nil {
					fmt.Fprintln(w, err)
				}
				template.Must(template.ParseFiles(indexPage)).Execute(w, bson.M{
					"Site": r.Host,
					"Error": bson.M{
						"NotFound": "Account not found!",
					},
				})
				return
			}

			if err := client.Disconnect(ctx); err != nil {
				fmt.Fprintln(w, err)
			}
			http.Redirect(w, r, fmt.Sprintf("/%s", existingUser), http.StatusSeeOther)
			return
		}

		userErrors, ok := user.IsValid()
		if !ok {
			template.Must(template.ParseFiles(indexPage)).Execute(w, bson.M{
				"Site":  r.Host,
				"Error": userErrors,
			})
			return
		}

		collection := getCollection(usersCollection, w)

		if err := collection.FindOne(ctx, &bson.M{"username": user.Username}).Err(); err != mongo.ErrNoDocuments {
			template.Must(template.ParseFiles(indexPage)).Execute(w, bson.M{
				"Site": r.Host,
				"Error": bson.M{
					"Username": "Username already taken!",
				},
			})
			return
		}

		res, err := collection.InsertOne(ctx, user)
		if err != nil {
			fmt.Fprintln(w, err.Error())
		}

		if err := client.Disconnect(ctx); err != nil {
			fmt.Fprintln(w, err)
		}

		template.Must(template.ParseFiles(indexPage)).Execute(w, bson.M{
			"Site":    r.Host,
			"Success": res.InsertedID.(primitive.ObjectID).Hex(),
		})
	})

	// Get all URLs for specific user
	r.Get("/{username}", func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var links []models.Link

		if err := getCollection(usersCollection, w).FindOne(ctx, bson.M{"username": chi.URLParam(r, "username")}).Decode(&user); err == mongo.ErrNoDocuments {
			if err := client.Disconnect(ctx); err != nil {
				fmt.Fprintln(w, err)
			}
			template.Must(template.ParseFiles(notFoundPage)).Execute(w, bson.M{
				"Site":    r.Host,
				"Message": "Account Not Found",
			})
			return
		}

		cursor, err := getCollection(urlsCollection, w).Find(ctx, bson.M{"createdBy": user.ID})
		if err != nil {
			template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
				"Site":     r.Host,
				"List":     false,
				"UserID":   user.ID.Hex(),
				"Username": user.Username,
			})
			return
		}

		cursor.All(ctx, &links)

		if err := client.Disconnect(ctx); err != nil {
			fmt.Fprintln(w, err)
		}

		for it, link := range links {
			link.FormatTime()
			links[it] = link
		}

		template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
			"Site":     r.Host,
			"UserID":   user.ID.Hex(),
			"Username": user.Username,
			"List":     links,
		})
	})

	// Create a new short URL
	r.Post("/{username}", func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var link models.Link
		var links []models.Link

		link.ID = primitive.NewObjectID()
		link.CreatedAt = time.Now()

		link.Url = r.PostFormValue("url")
		link.Info = r.PostFormValue("info")

		parsedTime, err := time.Parse("2006-01-02T15:04", r.PostFormValue("expiry"))
		if err != nil {
			fmt.Fprintln(w, err.Error())
		}
		link.ExpireAt = parsedTime

		if err := getCollection(usersCollection, w).FindOne(ctx, &bson.M{"username": chi.URLParam(r, "username")}).Decode(&user); err == mongo.ErrNoDocuments {
			template.Must(template.ParseFiles(notFoundPage)).Execute(w, bson.M{
				"Message": "Account not found!",
			})
			return
		}

		cursor, err := getCollection(urlsCollection, w).Find(ctx, bson.M{"createdBy": user.ID})
		if err != nil {
			template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
				"Site":     r.Host,
				"Path":     r.URL,
				"List":     false,
				"UserID":   user.ID.Hex(),
				"Username": user.Username,
			})
			return
		}

		cursor.All(ctx, &links)

		for it, link := range links {
			link.FormatTime()
			links[it] = link
		}

		linkErrors, ok := link.IsValid(user.Password, r.PostFormValue("password"))
		if !ok {
			template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
				"Site":     r.Host,
				"Path":     r.URL,
				"UserID":   user.ID.Hex(),
				"Username": user.Username,
				"Error":    linkErrors,
				"List":     links,
			})
			return
		}

		count, err := getCollection(urlsCollection, w).CountDocuments(ctx, &bson.M{"url": link.Url, "createdBy": user.ID})
		if err != nil {
			fmt.Fprintln(w, err.Error())
		}

		if count > 0 {
			template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
				"Site":     r.Host,
				"Path":     r.URL,
				"UserID":   user.ID.Hex(),
				"Username": user.Username,
				"Error": models.LinkError{
					Url: "Url already saved!",
				},
				"List": links,
			})
			return
		}

		link.CreatedBy = user.ID
		link.Short = RandomString(5)

		res, err := getCollection(urlsCollection, w).InsertOne(ctx, link)
		if err != nil {
			fmt.Fprintln(w, err.Error())
		}

		template.Must(template.ParseFiles(linksPage)).Execute(w, bson.M{
			"Site":     r.Host,
			"Path":     r.URL,
			"UserID":   user.ID.Hex(),
			"Username": user.Username,
			"Success":  res.InsertedID.(primitive.ObjectID).Hex(),
			"List":     append(links, link),
		})
	})

	// Direct url redirect
	r.Get("/{username}/{link}", func(w http.ResponseWriter, r *http.Request) {
		var link models.Link
		var user models.User

		if err := getCollection(usersCollection, w).FindOne(ctx, &bson.M{"username": chi.URLParam(r, "username")}).Decode(&user); err == mongo.ErrNoDocuments {
			template.Must(template.ParseFiles(notFoundPage)).Execute(w, bson.M{
				"Site":    r.Host,
				"Message": "User not found!",
			})
			return
		}

		if err := getCollection(urlsCollection, w).FindOne(ctx, &bson.M{"short": chi.URLParam(r, "link"), "createdBy": user.ID}).Decode(&link); err == mongo.ErrNoDocuments {
			template.Must(template.ParseFiles(notFoundPage)).Execute(w, bson.M{
				"Site":    r.Host,
				"Message": "URL Not Found!",
			})
			return
		}

		if err := client.Disconnect(ctx); err != nil {
			fmt.Fprintln(w, err)
		}

		http.Redirect(w, r, link.Url, http.StatusFound)
	})

	// Start server
	if err := http.ListenAndServe(":8090", r); err != nil {
		panic("Error starting server!")
	}
}
