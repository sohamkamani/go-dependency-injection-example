# The many flavours of dependency injection in Go

One of the most challenging aspects of building applications in Go is managing the many dependencies that your application uses. In an ideal world, your application would be stateless, with only one input and one output -- essentially acting like a pure function.

However, most medium to large scale applications _will_ have at least some dependencies. The applications we build at Gojek almost always have more than one of the following:

- A postgres database
- A redis cache
- An HTTP client
- A message queue
- Another HTTP client

For each of these dependencies, there is a bunch of stuff we have to consider:
- __Initialization__: How are we going to set up the initial connection or state of a dependency? This is something that will need to happen only once in the applications life cycle.
- __Testing__: How can we write independant test cases for services using an external dependency? Keep in mind that while writing test cases, we need to simulate the failure of our dependencies as well.
- __State__: How do we expose a reference to each dependency (which is supposed to be constant) without creating any sort of global state (which, as we all know, is the root of all evil)?

## Dependency injection to the rescue

The dependency injection pattern helps us solve these problems. Treating the external dependencies of our applications as individual dependencies for each service in our codebase allows us to have a more modular, and focussed approach during development, and follows a few premises:

- __Dependencies are stateful__ : The only reason you would consider treating something as a dependency is if it had some sort of state. For example, a database has its connection pool as its state. This also means that there should be some kind of initialization involved before you can use a dependency
- __Dependencies are represented by interfaces__ : A dependency is characterized by its contract. The service using it should not know about its implementation or internal state.

## Using dependency injection in a Go application

We can illustrate the use of dependency injection by building an application that makes use of it. Let's consider a service that is dependent on a database as its store. The service will fetch an entry from a database, and log the result after performing some validations.

We can define the service as a structure with the store as its dependency:

```go
package service

type Service struct {
	store database.Store
}
```

Here, `database.Store` is the dependencies interface, that we can define in another package:

```go
package database

type Store interface {
  // Get will fetch the value (which is an integer) for a given ID
	Get(ID int) (int, error)
}
```

The service can then use the dependency through its methods:

```go
func (s *Service) GetNumber(ID int) error {
	// Use the `Get` method of the dependency to retreive the value of the database entry
	result, err := s.store.Get(ID)
	if err != nil {
		return err
	}
	// Perform some validation, and output an error if it is too high
	if result > 10 {
		return fmt.Errorf("result too high: %d", result)
	}
	// Return nil, if the result is valid
	return nil
}
```

Note, that we have not defined the implementation of our dependency as yet (in fact, that's one of the last things we will do!)

### Testing the service

One of the most powerful features of dependency injection, is that you can test any dependant service without having any actual implementation of the dependency. In fact, we can mock our dependency to behave the way we want it to, so that we can test our service to handle different failure scenarios.

First, we will have to mock our dependency:

```go
package database

import (
  // We use the "testify" library for mocking our store
	"github.com/stretchr/testify/mock"
)

// Create a MockStore struct with an embedded mock instance
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(ID int) (int, error) {
	// This allows us to pass in mocked results, so that the mock store will return whatever we define 
  returnVals := m.Called(ID)
  // return the values which we define
	return returnVals.Get(0).(int), returnVals.Error(1)
}
```

We can then use this mock store to simulate the dependency in our service when we test it:

```go
func TestServiceSuccess(t *testing.T) {
	// Create a new instance of the mock store
	m := new(database.MockStore)
	// In the "On" method, we assert that we want the "Get" method
	// to be called with one argument, that is 2
	// In the "Return" method, we define the return values to be 7, and nil (for the result and error values)
	m.On("Get", 2).Return(7, nil)
	// Next, we create a new instance of our service with the mock store as its "store" dependency
	s := Service{m}
	// The "GetNumber" method call is then made
	err := s.GetNumber(2)
	// The expectations that we defined for our mock store earlier are asserted here
	m.AssertExpectations(t)
	// Finally, we assert that we should'nt get any error
	if err != nil {
		t.Errorf("error should be nil, got: %v", err)
	}
}
```

We can now use the mock to simulate error scenarios, and test for them as well:

```go
func TestServiceResultTooHigh(t *testing.T) {
	m := new(database.MockStore)
	// In this case, we simulate a return value of 24, which would fail the services validation
	m.On("Get", 2).Return(24, nil)
	s := Service{m}
	err := s.GetNumber(2)
	m.AssertExpectations(t)
	// We assert that we expect the "result too high" error given by the service
	if err.Error() != "result too high: 24" {
		t.Errorf("error should be 'result too high: 24', got: %v", err)
	}
}

func TestServiceStoreError(t *testing.T) {
	m := new(database.MockStore)
	// In this case, we simulate the case where the store returns an error, which may occur if it is unable to fetch the value
	m.On("Get", 2).Return(0, errors.New("failed"))
	s := Service{m}
	err := s.GetNumber(2)
	m.AssertExpectations(t)
	if err.Error() != "failed" {
		t.Errorf("error should be 'failed', got: %v", err)
	}
}
```

### Implementing and initializing the real store

Now that we know our service works well with the mock store, we can implement the actual one:

```go
// The actual store would contain some state. In this case it's the sql.db instance, that holds the connection to our database
type store struct {
	db *sql.DB
}

// Implement the "Get" method, in order to comply with the "Store" interface
func (d *store) Get(ID int) (int, error) {
  //we would perform some external database operation with d.db
  // for the sake of clarity, that code is not shown here
	return 0, nil
}

// Add a constructor function to return a new instance of a store
func NewStore(db *sql.DB) Store {
	return &store{db}
}
```

We can now put together the "store" as a dependency to the service and construct a simple command line app:

```go
func main() {
	// Create a new DB connection
	connString := "dbname=<your main db name> sslmode=disable"
	db, _ := sql.Open("postgres", connString)

	// Create a store dependency with the db connection
	store := database.NewStore(db)
	// Create the service by injecting the store as a dependency
	service := &service.Service{Store: store}

	// The following code implements a simple command line app to read the ID as input
	// and output the validity of the result of the entry with that ID in the database
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(){
		ID, _ := strconv.Atoi(scanner.Text())
		err := service.GetNumber(ID)
		if err != nil {
			fmt.Printf("result invalid: %v", err)
			continue
		}
		fmt.Println("result valid")
	}
}
```

What we have essentially done by using dependency injection, is converted something that looks like a dependency graph, into something that looks like a pure function, with the dependencies now part of the service:


## Alternative implementations of dependency injection

Adding dependencies as object attributes isn't the only way to inject them. Sometimes, when your interface has just one method, it's more convenient to use a more functional form of dependency injection. If we were to assume that our service _only_ had a `GetNumber` method, we could use curried functions to add dependencies as variables inside the closure:

```go
func NewGetNumber(store database.Store) func(int) error {
	return func(ID int) error {
		// "store" is still a dependency, only, now it's accessible through the function closure
		result, err := store.Get(ID)
		if err != nil {
			return err
		}
		if result > 10 {
			return fmt.Errorf("result too high: %d", result)
		}
		return nil
	}
}
``` 

And you can generate the `GetNumber` function by calling the function constructor with a store implementation:

```go
GetNumber := NewGetNumber(store)
```

`GetNumber` now has the same functionality as the previous OOP based approach. This method of deriving dependant functions using currying is especially useful when you require single functions instead of a whole suite of methods (for example, in HTTP handler functions).

## When to avoid dependency injection

As with all things, there is no silver bullet to solve all your problems, and this is true for the dependency injection pattern as well. Although it can make your code more modular, this pattern also comes with the cost of increased complexity during initialization. You cannot simply call a method of a dependancy without explicitly passing it down during initialization. This also makes it harder to add new services, since there is more boilerplate code to get it up and running the first time. Sometimes, when there are a lot of embedded dependencies (if your dependencies have their own dependencies), initialization can be a nightmare.

If the application you are building is simple, or if you have many embedded dependencies then you should probably assess if it is worth the trade-offs to implement this pattern.