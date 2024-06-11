package parser

// Parser defines a generic interface for parsing HTML content.
// This interface is designed to be implemented by various types of parsers, each tailored to handle specific HTML
// structures from different websites.
//
// Type parameter:
//   - T any: The type of the result produced by the parser. It allows the parser to be used in a type-safe manner
//     with different data structures or models, ensuring that the consumer of the parser knows exactly what type
//     of data to expect in return.
//
// Methods:
//   - Parse: Takes a string containing HTML content as input and processes it to return a pointer to a result of type T
//     and an error if the processing fails. This method is meant to encapsulate all necessary steps for converting
//     the input HTML string into a more structured form of type T. It is expected that implementations of this interface
//     handle all errors internally and return an error only if it is not possible to recover from it or if the error
//     needs to be communicated to the caller.
type Parser[T any] interface {
	// Parse takes a string of HTML content as input and attempts to process it into a structured form of type T.
	// The method returns a pointer to the result of type T, which represents the processed data, and an error if
	// the processing fails.
	//
	// Parameters:
	// - htmlContent string: The raw HTML content that needs to be processed. This could be any form of HTML text,
	//                       depending on the specific implementation of the parser.
	//
	// Returns:
	// - *T: A pointer to the result of type T, which represents the structured data extracted from the HTML content.
	// - error: An error object that encapsulates any issue that occurred during the parsing. If the parsing is
	//          successful, this will be nil.
	Parse(htmlContent string) (*T, error)
}
