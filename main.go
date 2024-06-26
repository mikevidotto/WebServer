package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//--------------INITIAL SETUP================

// 1. add Post struct
type Post struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// 2. add global variables
var (
	posts   = make(map[int]Post)
	nextID  = 1
	postsMu sync.Mutex
)

//--------------IMPLEMENTING SERVER================

// 3. add HandleFuncs and start server listening at localhost.
func main() {

	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/posts/", postHandler)

	fmt.Println("Server is running at the http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

//--------------HANDLING REQUESTS================

// 4. postsHandler function
func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetPosts(w, r)
	case "POST":
		handlePostPosts(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 5. postHandler function
func postHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/posts/"):])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleGetPost(w, r, id)
	case "DELETE":
		handleDeletePost(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

//--------------CRUD OPERATIONS================

func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// this essentially locks the server so that we can
	// manipulate the posts map without worrying about
	// another request trying to do the same thing at
	// the same time.
	postsMu.Lock()

	// defers unlocking until the function has finished executing,
	// but define it up the top with our lock. Nice and neat.
	// Caution: deferred statements are first-in-last-out,
	// which is not all that intuitive to begin with.
	defer postsMu.Unlock()

	// Copying the posts to a new slice of type []Post
	ps := make([]Post, 0, len(posts))
	for _, p := range posts {
		ps = append(ps, p)
	}

	fmt.Println(ps)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

func handlePostPosts(w http.ResponseWriter, r *http.Request) {
	var p Post

	// This will read the entire body into a byte slice
	// i.e. ([]byte)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Now we'll try to parse the body. This is similar
	// to JSON.parse in JavaScript.
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	// As we're going to mutate the posts map, we need to
	// lock the server again
	postsMu.Lock()
	defer postsMu.Unlock()

	p.ID = nextID
	nextID++
	posts[p.ID] = p

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func handleGetPost(w http.ResponseWriter, r *http.Request, id int) {
	postsMu.Lock()
	defer postsMu.Unlock()

	p, ok := posts[id]
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func handleDeletePost(w http.ResponseWriter, r *http.Request, id int) {
	postsMu.Lock()
	defer postsMu.Unlock()

	// If you use a two-value assignment for accessing a
	// value on a map, you get the value first then an
	// "exists" variable.
	_, ok := posts[id]
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	delete(posts, id)
	w.WriteHeader(http.StatusOK)
}
