package log

import (
	"log"
	"os"
)

var (
	logger *log.Logger
	errLogger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", 0)
	errLogger = log.New(os.Stderr, "\033[31mERROR\033[0m ", 0)
}

func Print(args ...interface{})  {
	logger.Print(args)
}

func Printf(format string, args ...interface{})  {
	logger.Printf(format, args...)
}

func Println(args ...interface{}) {
	logger.Println(args...)
}

func Fatal(args ...interface{}) {
	errLogger.Fatal(args)
}

func Fatalf(format string, args ...interface{})  {
	errLogger.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	errLogger.Fatalln(args...)
}

func Error(args ...interface{}) {
	errLogger.Print(args)
}

func Errorf(format string, args ...interface{})  {
	errLogger.Printf(format, args...)
}

func Errorln(args ...interface{}) {
	errLogger.Println(args...)
}

func Panic(args ...interface{}) {
	errLogger.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	errLogger.Panicf(format, args...)
}

func Panicln(args ...interface{}) {
	errLogger.Panicln(args...)
}