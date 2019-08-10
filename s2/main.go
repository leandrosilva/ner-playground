package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	prose "gopkg.in/jdkato/prose.v2"
)

//-- DATA: LOAD AND SPLIT

type labeledEntities struct {
	Text   string
	Spans  []prose.LabeledEntity
	Answer string
}

func readFile(fileName string) []byte {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	return data
}

func mapData(jsonl []byte) []labeledEntities {
	decoder := json.NewDecoder(bytes.NewReader(jsonl))
	entries := []labeledEntities{}

	for {
		entity := labeledEntities{}
		err := decoder.Decode(&entity)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		entries = append(entries, entity)
	}

	return entries
}

func mapEntities(data []labeledEntities) []prose.EntityContext {
	entities := []prose.EntityContext{}
	for _, entity := range data {
		entities = append(entities, prose.EntityContext{
			Text:   entity.Text,
			Spans:  entity.Spans,
			Accept: entity.Answer == "accept"})
	}

	return entities
}

func splitEntities(entities []prose.EntityContext, trainSplit float64) ([]prose.EntityContext, []prose.EntityContext) {
	trainSize := int(float64(len(entities)) * trainSplit)

	train, test := []prose.EntityContext{}, []prose.EntityContext{}
	for i, entity := range entities {
		if i < trainSize {
			train = append(train, entity)
		} else {
			test = append(test, entity)
		}
	}

	return train, test
}

//-- MODEL: TRAIN AND TEST

func createModel(name string, train []prose.EntityContext) *prose.Model {
	model := prose.ModelFromData(name, prose.UsingEntities(train))
	return model
}

func saveModelToDisk(name string, model *prose.Model) {
	os.RemoveAll("./" + name + "/Maxent")
	model.Write(name)
}

func loadModelFromDisk(name string) *prose.Model {
	model := prose.ModelFromDisk(name)
	return model
}

func testModel(model *prose.Model, test []prose.EntityContext) {
	n := len(test)
	correct := 0.0

	for _, entity := range test {
		doc, err := prose.NewDocument(entity.Text, prose.WithSegmentation(false), prose.UsingModel(model))
		if err != nil {
			panic(err)
		}
		entities := doc.Entities()

		if !entity.Accept && len(entities) == 0 {
			// If it was rejected before, it is labeled, so ignore it and go ahead
			correct++
		} else {
			expected := []string{}
			for _, span := range entity.Spans {
				expected = append(expected, entity.Text[span.Start:span.End])
			}

			if reflect.DeepEqual(expected, entities) {
				correct++
			}
		}
	}

	fmt.Println("Test Model ( n =", n, ", correct =", correct, ", % =", (correct / float64(n)), ")")
}

//-- ENTITY RECOGNITION

type entityCount struct {
	Entity prose.Entity
	Count  int
}

func distinctEntities(entities []prose.Entity) []entityCount {
	counter := map[string]entityCount{}
	for _, entity := range entities {
		value, found := counter[entity.Text]
		if found {
			value.Count = value.Count + 1
			counter[entity.Text] = value
		} else {
			counter[entity.Text] = entityCount{Entity: entity, Count: 1}
		}
	}

	distinct := []entityCount{}
	for _, value := range counter {
		distinct = append(distinct, value)
	}

	return distinct
}

func recognizeEntity(text string, model *prose.Model, printText bool) {
	doc, err := prose.NewDocument(text, prose.WithSegmentation(false), prose.UsingModel(model))
	if err != nil {
		panic(err)
	}

	fmt.Println("Recognizing entity...")
	if printText {
		fmt.Println(">", text)
	}
	if len(doc.Entities()) == 0 {
		fmt.Println("( No named entity was recognized. )")
	} else {
		for _, ec := range distinctEntities(doc.Entities()) {
			fmt.Println("( POS Tag:", ec.Entity.Text, ", IOB Label:", ec.Entity.Label, ", Count:", ec.Count, ")")
		}
	}
}

//-- BOOT

func main() {
	shouldTrainModel := flag.Bool("train", false, "Should train model or load from disk")
	flag.Parse()

	fmt.Println("Getting data...")
	entities := mapEntities(mapData(readFile("reddit_product.jsonl")))
	train, test := splitEntities(entities, 0.8)

	modelName := "PRODUCT"
	var model *prose.Model

	if *shouldTrainModel {
		fmt.Println("Training entities...")
		start := time.Now()
		model = createModel(modelName, train)
		fmt.Println("Train:", len(train), "Test:", len(test), "in", time.Now().Sub(start))

		fmt.Println("Saving model to disk...")
		saveModelToDisk(modelName, model)
	} else {
		fmt.Println("Loading model from disk...")
		model = loadModelFromDisk(modelName)
	}

	testModel(model, test)

	recognizeEntity("Well, Windows 10 is not a Mac OSX or Linux but it is not that bad.", model, true)
	recognizeEntity("Who in the freakin' Earth would use Bing instead of Google?", model, true)
	recognizeEntity("Look, guy let his iPhone just there.", model, true)
	recognizeEntity("Would it that whole pizza, bro.", model, true)
	recognizeEntity(string(readFile("big_text.txt")), model, false)
}
