package database

func MongoConnect() string {
	return "teste"
}

// func MongoConnect() *mongo.Client {
// 	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
// 	        err := godotenv.Load()
// 	        if err != nil {
// 	            log.Println("Arquivo .env não encontrado, seguindo com variáveis de ambiente do sistema.")
// 		}
// 	}

// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
// 	opts := options.Client().ApplyURI(fmt.Sprintf("mongodb+srv://%s:%s@cluster0.ehrx4zi.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))).SetServerAPIOptions(serverAPI)
// 	client, err := mongo.Connect(context.TODO(), opts)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
// 	return client
// }
