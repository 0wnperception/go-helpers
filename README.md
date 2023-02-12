# PROGRAMS #
## RUN ##
`go run .`
run program

# MAC CROSS COMPLILING #
GOOS=linux GOARCH=arm go build -o bin_name .

## BUILD ##
`go build . -o 'name'`
build program

### OS LIST ###
- OS	$GOOS
- Linux	`linux`
- MacOS X	`darwin`
- Windows	`windows`
- FreeBSD	`freebsd`
- NetBSD	`netbsd`
- OpenBSD	`openbsd`
- DragonFly BSD	`dragonfly`
- Plan 9	`plan9`
- Native Client	`nacl`
- Android	`android`

### ARCH ###
- Architecture	$GOARCH
- x386	`386`
- AMD64	`amd64`
- AMD64 с 32-указателями	`amd64p32`
- ARM	`arm`

# MODULES #
- `go mod init 'moule name'`
- `go clean --modcache`
init module
- `go work use .`

# GENERATORS #
`go generate .`
generate all options in .
### MOCKS ###
`//go:generate minimock -i path/repo.Iface -o ./tests/_mock.go -n _mock`

# LINTER (golangci-lint) #
- `golangci-lint help linters`
show all linters
- `golangci-lint run --disable-all`
disable all linters
- `golangci-lint run --enable-all`
enable all linters
- `golangci-lint run -E 'linter name'`
enable linter
- `golangci-lint run -D 'linter name'`
disable linter

# TESTS #
https://golang-blog.blogspot.com/2019/07/testing-flags.html

## RUN ##
`go test . -`

## CACHE OFF ## 
- `GOCACHE=off go test ./... -test.v`
- `go clean -testcache`

## FLAGS ##
- `-run 'test name'`
set special test name 
- `-v`
visual stdout
- `-bench=.`
Запустить бенчмарки
- `-benchmem`
Распечатать статистику распределения памяти для тестов.
- `-blockprofile block.out`
Записать профиль блокировки goroutine в указанный файл когда все тесты завершены. Записывает тестовый бинарный файл как делает флаг -c. 
- `-coverprofile cover.out`
Записать профиль покрытия в файл после того, как все тесты пройдены.
Включает флаг -cover.
- `-cpuprofile cpu.out`
Записать профиль CPU в указанный файл перед выходом. Записывает тестовый бинарный файл как делает флаг -c.
- `-memprofile mem.out`
Записать профиль распределения в файл после того, как все тесты пройдены. Записывает тестовый бинарный файл как флаг -c.
- `-memprofilerate n`
Включить более точные (и дорогие) профили выделения памяти установкой runtime.MemProfileRate. Смотрите 'go doc runtime.MemProfileRate'. Чтобы профилировать все распределения памяти, используйте -test.memprofilerate=1.
- `-mutexprofile mutex.out`
Записать конфликтный профиль мьютекса в указанный файл когда все тесты завершены. Записывает тестовый бинарный файл как флаг -c.
- `-trace trace.out`
Записать трассировку выполнения в указанный файл перед выходом.

## OPEN METRICS ##
- go tool pprof ./block.out
- go tool pprof *.test ./cpu.out
- go tool pprof ./mem.out
- go tool pprof ./mutex.out
- go tool trace ./trace.out
- go tool cover -func=cover.out
- go tool cover -html=cover.out

- pprof -http=:6060

### PPROF OPTIONS ###
- top
- web
- list (func name)
- sample_index = 


# Shuduler #
https://habr.com/ru/post/478168/
https://habr.com/ru/post/489862/
- GOMAXPROC set virtual processors to use

## TAGS ##
- //go:nosplit set to not interrupt runtime of gorutine

# GC #
https://www.youtube.com/watch?v=CX4GSErFenI
- GOGC - percentage of memory to start garbage collector

# chan #
https://habr.com/ru/post/308070/