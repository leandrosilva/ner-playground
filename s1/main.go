package main

import (
	"fmt"

	prose "gopkg.in/jdkato/prose.v2"
)

func main() {
	text := "European authorities fined Google a record $5.1 billion on Wednesday for abusing its power in the mobile phone market and ordered the company to alter its practices"
	fmt.Println("Text:", text)

	// The document-creation process consists of four steps: tokenization, segmentation,
	// POS tagging, and named-entity extraction, which can be disabled by WithTokenization,
	// WithSegmentation, WithTagging, and WithExtraction functions.
	//
	// Let's go with default here.
	sentence, err := prose.NewDocument(text)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nChunking...\nToken: Text >> Part-Of-Speech Tag >> IOB Label")
	for _, token := range sentence.Tokens() {
		fmt.Println("(", token.Text, ", ", token.Tag, ", ", token.Label, ")")
	}

	fmt.Println("\nNamed Entity...\nToken: Text >> Part-Of-Speech Tag >> IOB Label")
	for _, entity := range sentence.Entities() {
		fmt.Println("(", entity.Text, ",", entity.Label, ")")
	}
	fmt.Println("\nFor the love of Zeus, where is Google?")

	fmt.Println("\nTraining a new model...")
	train := []prose.EntityContext{
		{
			Text:   "Google is a gigantic international company",
			Spans:  []prose.LabeledEntity{{Start: 0, End: 5, Label: "GPE"}},
			Accept: true},
		{
			Text:   "Google releases a brand new product every so often",
			Spans:  []prose.LabeledEntity{{Start: 0, End: 5, Label: "GPE"}},
			Accept: true},
		{
			Text:   "There is no company bigger than Google in this internet",
			Spans:  []prose.LabeledEntity{{Start: 32, End: 38, Label: "GPE"}},
			Accept: true}}
	model := prose.ModelFromData("Google", prose.UsingEntities(train))
	newSentence, _ := prose.NewDocument(text, prose.UsingModel(model))

	fmt.Println("Named Entity after some training...\nToken: Text >> Part-Of-Speech Tag >> IOB Label")
	for _, entity := range newSentence.Entities() {
		fmt.Println("(", entity.Text, ",", entity.Label, ")")
	}
	fmt.Println("\nOK. Right there.")
}
