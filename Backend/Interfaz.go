package Backend

import "fmt"

func main() {
	category := SelecRandomCategory()

	questionlist, err := CreateQuestionList(category)

	if err != nil {
		fmt.Println("Error reading the file:", err)
		return
	}

	GiveRandomQuestionToPlayer(questionlist)

}
