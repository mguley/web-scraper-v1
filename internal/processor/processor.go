package processor

// Processor defines a generic interface for processing payloads.
// This interface is designed to be implemented by various types of processors, each tailored to handle
// specific kinds of data or operations, such as parsing, validation, transformation, or storage tasks.
//
// Type parameter:
//   - T any: The type of the result produced by the processor. It allows the processor to be used in a type-safe
//     manner with different data structures or models, ensuring that the consumer of the processor knows
//     exactly what type of data to expect in return.
//
// Methods:
//   - Process: Takes a string as input and processes it to return a pointer to a result of type T and an error if the
//     processing fails. This method is meant to encapsulate all necessary steps for converting the input string
//     into a more structured form of type T. It is expected that implementations of this interface handle all
//     errors internally and return an error only if it is not possible to recover from it or if the error needs
//     to be communicated to the caller.
type Processor[T any] interface {
	// Process takes a string item as input and attempts to process it into a structured form of type T.
	// The method returns a pointer to the result of type T, which represents the processed data, and an error
	// if the processing fails.
	//
	// Parameters:
	// - item string: The raw string data that needs to be processed. This could be any form of text, such as JSON,
	//                XML, plain text, etc., depending on the specific implementation of the processor. If the processing
	//                is unsuccessful, this will be nil.
	// - error: An error object that encapsulates any issue that occurred during the processing. If the processing is
	//          successful, this will be nil.
	Process(item string) (*T, error)
}
