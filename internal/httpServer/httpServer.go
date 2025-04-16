package httpServer

import (
  "log"
  "net/http"
  "strconv"
  "encoding/json"
  "sync"
  //"os"
  "time"
  "fmt"
  "html/template"
  "context"
  "io/ioutil"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
)
type Server struct{
	collection  *mongo.Collection
}

type Request struct{
	Text string `json:"text"`
}

type NewBill struct{
	Name string `json:"name"`
	Value string `json:"value"`
}

type Record struct {
    ID    primitive.ObjectID `bson:"_id" json:"id"`
    Name  string `bson:"name" json:"name"`
	Role  string `bson:"role" json:"role"`
    Value int `bson:"value" json:"value"`
}

type ResponseRes struct{
	Result string `json:"result"`
}

type StatsResp struct
{
	BillsCount int64		`json:"bills"`
	EmployeesCount int64	`json:"people"`
}

type Fine struct {
	ID    primitive.ObjectID `bson:"_id" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Role  string             `bson:"role" json:"role"`
	Value int                `bson:"value" json:"value"`
}

type Response struct{
	Result []string `json:"result"`
}

type Name struct{
	Name string `json:"name"`
}

func stringifyFine(fine Fine) string {
	return fmt.Sprintf("Bill with id: %v for %s with role '%s', in amount of %v.", fine.ID, fine.Name, fine.Role, fine.Value)
}

func (s *Server)finesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL)

	body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error while getting data", err)
        return
    }
	log.Println("body", string(body))

	var resp Name
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Printf("Error while converting body: %s", err)
		return
	}

	log.Println(resp.Name)
    
    
    // Формирование фильтра
	filter := bson.M{
		"$where": fmt.Sprintf("this.name === '%s'", resp.Name),
	}

    log.Println("Filter", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение запроса без проекции
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Error in collection.Find: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []Fine
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Error in cursor.All: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var message []string

	for _, fine := range results{
		message = append(message, stringifyFine(fine))
	}
	log.Println(message)
	// Отправка полных результатов (включая role и value)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: message})
}

func (s *Server)homeHandler(w http.ResponseWriter, r *http.Request) {
	//path := filepath.Join("./templates", "index.html")

	t, err := template.ParseFiles("../../templates/index.html")
	if err != nil{
		http.Error(w, err.Error(), 400)
		log.Printf("Failed to make html page: %v", err)
	}
	err = t.Execute(w, nil)
	if err != nil{
		http.Error(w, err.Error(), 400)
		log.Printf("Failed to make html page: %v", err)
	}
}

func (s *Server)addBill(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Printf("Trying to add a bill")
	var input NewBill
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error while getting data", err)
        return
    }

	name := input.Name
	value, _ := strconv.Atoi(input.Value)
	if val, _ := strconv.Atoi(input.Value); val <= 0{
		log.Printf("Wrong data given...")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponseRes{Result: "You entered wrong data"})
		return
	}


	newRecord := Record{
        ID:    primitive.NewObjectID(), // Генерация уникального ID
        Name:  name,
        Value: value,
        Role:  "user", // Фиксированное значение
    }

	_, err := s.collection.InsertOne(ctx, newRecord)
	if err != nil{
		log.Println(err)
	}
	log.Printf("END")

	w.Header().Set("Content-Type", "application/json")
	if err == nil{
		json.NewEncoder(w).Encode(ResponseRes{Result: "Succesfully added a bill!"})
	} else{
		json.NewEncoder(w).Encode(ResponseRes{Result: "Failed to add a bill. Try again..."})
	}
	
}

func (s *Server)checkstats(w http.ResponseWriter, r *http.Request){
	log.Printf("Got checkStats request")
	
	total, namesCount := countData(s.collection)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResp{BillsCount: total, EmployeesCount: namesCount})
}

func countData(collection *mongo.Collection) (int64, int64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

    distinctValues, err := collection.Distinct(ctx, "name", bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	count, err := collection.EstimatedDocumentCount(ctx)
    if err != nil {
        log.Fatal(err)
    }

    return count, int64(len(distinctValues))
}
	

func Start(wg *sync.WaitGroup, collection *mongo.Collection) {
	defer wg.Done()
	server := Server{collection: collection}
	//port := os.Getenv("PORT")
	port := ":8080"
	http.HandleFunc("/getstats", server.checkstats)
	http.HandleFunc("/bills/add", server.addBill)
	http.HandleFunc("/bills/check", server.finesHandler)
	http.HandleFunc("/", server.homeHandler)
	log.Println("Starting server on 8080...")
	log.Fatal(http.ListenAndServe(port, nil))
}