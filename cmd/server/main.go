package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Структуры данных (в памяти)
type Store struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Logo      string    `json:"logo"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Photo       string    `json:"photo"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

// Хранилище в памяти
var (
	stores    = make(map[int]Store)
	products  = make(map[int]Product)
	users     = make(map[int]User)
	storeID   = 1
	productID = 1
	mu        sync.RWMutex
)

func init() {
	// Инициализируем тестовые данные
	initTestData()
}

func initTestData() {
	// Тестовые магазины
	stores[1] = Store{
		ID:        1,
		Name:      "Продуктовый рай",
		Address:   "ул. Ленина, 10",
		Logo:      "/static/images/store1.jpg",
		CreatedAt: time.Now(),
	}
	stores[2] = Store{
		ID:        2,
		Name:      "Супермаркет 'У дома'",
		Address:   "пр. Мира, 25",
		Logo:      "/static/images/store2.jpg",
		CreatedAt: time.Now(),
	}

	// Тестовые товары
	products[1] = Product{
		ID:          1,
		Name:        "Молоко",
		Price:       89.90,
		Description: "Свежее молоко 3.2%",
		Photo:       "/static/images/milk.jpg",
		CreatedAt:   time.Now(),
	}
	products[2] = Product{
		ID:          2,
		Name:        "Хлеб",
		Price:       45.50,
		Description: "Белый хлеб нарезной",
		Photo:       "/static/images/bread.jpg",
		CreatedAt:   time.Now(),
	}

	// Тестовый пользователь
	users[1] = User{
		ID:       1,
		Username: "admin",
		Password: "admin123",
	}
}

func main() {
	// Настройка Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Загружаем HTML шаблоны
	r.LoadHTMLGlob("templates/*.html")

	// Статические файлы
	r.Static("/static", "./static")

	// Маршруты
	setupRoutes(r)

	// Запуск сервера
	port := ":8080"
	log.Printf("🚀 Сервер запущен на http://localhost%s", port)

	// Пробуем разные порты если 8080 занят
	err := r.Run(port)
	if err != nil {
		log.Printf("Порт 8080 занят, пробуем 8081...")
		r.Run(":8081")
	}
}

func setupRoutes(r *gin.Engine) {
	// Главная страница
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	// Авторизация
	r.GET("/login", loginPage)
	r.POST("/login", loginHandler)
	r.GET("/logout", logoutHandler)

	// Магазины
	r.GET("/stores", storesList)
	r.GET("/stores/create", storeCreatePage)
	r.POST("/stores/create", storeCreate)
	r.GET("/stores/:id/edit", storeEditPage)
	r.POST("/stores/:id/edit", storeUpdate)
	r.GET("/stores/:id/delete", storeDelete)

	// Товары
	r.GET("/products", productsList)
	r.GET("/products/create", productCreatePage)
	r.POST("/products/create", productCreate)
	r.GET("/products/:id/edit", productEditPage)
	r.POST("/products/:id/edit", productUpdate)
	r.GET("/products/:id/delete", productDelete)

	// Статические страницы
	r.GET("/about", aboutPage)
	r.GET("/contacts", contactsPage)
	r.GET("/navigator", navigatorPage)
}

// ========== АВТОРИЗАЦИЯ ==========
func loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "admin" && password == "admin123" {
		// Устанавливаем сессию (упрощенно)
		c.SetCookie("auth", "true", 3600, "/", "", false, true)
		c.SetCookie("username", username, 3600, "/", "", false, false)

		c.Redirect(http.StatusFound, "/stores")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": "Неверный логин или пароль",
	})
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func checkAuth(c *gin.Context) bool {
	auth, err := c.Cookie("auth")
	return err == nil && auth == "true"
}

// ========== МАГАЗИНЫ ==========
func storesList(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	mu.RLock()
	storeList := make([]Store, 0, len(stores))
	for _, store := range stores {
		storeList = append(storeList, store)
	}
	mu.RUnlock()

	c.HTML(http.StatusOK, "stores.html", gin.H{
		"stores":      storeList,
		"total_count": len(storeList),
	})
}

func storeCreatePage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "store_form.html", gin.H{
		"title":  "Добавить магазин",
		"action": "/stores/create",
	})
}

func storeCreate(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	name := c.PostForm("name")
	address := c.PostForm("address")

	if name == "" {
		c.HTML(http.StatusOK, "store_form.html", gin.H{
			"title":  "Добавить магазин",
			"action": "/stores/create",
			"error":  "Название обязательно",
			"store":  Store{Name: name, Address: address},
		})
		return
	}

	mu.Lock()
	storeID++
	stores[storeID] = Store{
		ID:        storeID,
		Name:      name,
		Address:   address,
		CreatedAt: time.Now(),
	}
	mu.Unlock()

	c.Redirect(http.StatusFound, "/stores")
}

func storeEditPage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/stores")
		return
	}

	mu.RLock()
	store, exists := stores[id]
	mu.RUnlock()

	if !exists {
		c.Redirect(http.StatusFound, "/stores")
		return
	}

	c.HTML(http.StatusOK, "store_form.html", gin.H{
		"title":  "Редактировать магазин",
		"action": fmt.Sprintf("/stores/%d/edit", id),
		"store":  store,
	})
}

func storeUpdate(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/stores")
		return
	}

	mu.Lock()
	store, exists := stores[id]
	if !exists {
		mu.Unlock()
		c.Redirect(http.StatusFound, "/stores")
		return
	}

	store.Name = c.PostForm("name")
	store.Address = c.PostForm("address")
	stores[id] = store
	mu.Unlock()

	c.Redirect(http.StatusFound, "/stores")
}

func storeDelete(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/stores")
		return
	}

	mu.Lock()
	delete(stores, id)
	mu.Unlock()

	c.Redirect(http.StatusFound, "/stores")
}

// ========== ТОВАРЫ ==========
func productsList(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	mu.RLock()
	productList := make([]Product, 0, len(products))
	for _, product := range products {
		productList = append(productList, product)
	}
	mu.RUnlock()

	// Рассчитываем среднюю цену
	var totalPrice float64
	for _, p := range productList {
		totalPrice += p.Price
	}
	avgPrice := 0.0
	if len(productList) > 0 {
		avgPrice = totalPrice / float64(len(productList))
	}

	c.HTML(http.StatusOK, "products.html", gin.H{
		"products":      productList,
		"total_count":   len(productList),
		"average_price": fmt.Sprintf("%.2f", avgPrice),
	})
}

func productCreatePage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "product_form.html", gin.H{
		"title":  "Добавить товар",
		"action": "/products/create",
	})
}

func productCreate(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	description := c.PostForm("description")

	if name == "" || priceStr == "" {
		c.HTML(http.StatusOK, "product_form.html", gin.H{
			"title":  "Добавить товар",
			"action": "/products/create",
			"error":  "Название и цена обязательны",
			"product": Product{
				Name:        name,
				Description: description,
			},
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.HTML(http.StatusOK, "product_form.html", gin.H{
			"title":  "Добавить товар",
			"action": "/products/create",
			"error":  "Неверный формат цены",
			"product": Product{
				Name:        name,
				Description: description,
			},
		})
		return
	}

	mu.Lock()
	productID++
	products[productID] = Product{
		ID:          productID,
		Name:        name,
		Price:       price,
		Description: description,
		CreatedAt:   time.Now(),
	}
	mu.Unlock()

	c.Redirect(http.StatusFound, "/products")
}

func productEditPage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	mu.RLock()
	product, exists := products[id]
	mu.RUnlock()

	if !exists {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	c.HTML(http.StatusOK, "product_form.html", gin.H{
		"title":   "Редактировать товар",
		"action":  fmt.Sprintf("/products/%d/edit", id),
		"product": product,
	})
}

func productUpdate(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	mu.Lock()
	product, exists := products[id]
	if !exists {
		mu.Unlock()
		c.Redirect(http.StatusFound, "/products")
		return
	}

	product.Name = c.PostForm("name")
	product.Description = c.PostForm("description")

	priceStr := c.PostForm("price")
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		product.Price = price
	}

	products[id] = product
	mu.Unlock()

	c.Redirect(http.StatusFound, "/products")
}

func productDelete(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/products")
		return
	}

	mu.Lock()
	delete(products, id)
	mu.Unlock()

	c.Redirect(http.StatusFound, "/products")
}

// ========== СТАТИЧЕСКИЕ СТРАНИЦЫ ==========
func aboutPage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "about.html", gin.H{})
}

func contactsPage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "contacts.html", gin.H{})
}

func navigatorPage(c *gin.Context) {
	if !checkAuth(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "navigator.html", gin.H{})
}
