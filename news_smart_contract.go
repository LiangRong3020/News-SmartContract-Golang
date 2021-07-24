
package main

// 引入相關套件
import (
        "bytes"
        "encoding/json"
        "fmt"
        "github.com/hyperledger/fabric/core/chaincode/shim"
        sc "github.com/hyperledger/fabric/protos/peer"
        "strings"
)

// 定義所需的資料格式
// Define the Smart Contract structure
type SmartContract struct {

}

// 設定必須的 init 與 invoke 方法
/*
 * The Init method *
 called when the Smart Contract "tuna-chaincode" is instantiated by the network
 * Best practice is to have any Ledger initialization in separate function
 -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
        return shim.Success(nil)
}


type News struct {
        NewsId string `json:"newsId"`
        Title string `json:"title"`
        ReleaseDatetime string `json:"releaseDatetime"`
        Author  []string `json:"author"`
        Newsurl	string `json:"Newsurl"`
        Content string `json:"content"`
        Origin  string `json:"origin"`
}

type Comments struct {
        CommentId string `json:"commentId"`
        NewsId string `json:"newsId"`
        Comment string `json:"comment"`
        CommentDatetime string `json:"commentDatetime"`
        DisplayName string `json:"displayName"`
        UserId string `json:"userId"`
}

type Users struct {
        UserId  string `json:"userId"`
        DisplayName string `json:"displayName"`
        PictureUrl string `json:"pictureUrl"`
        StatusMessage string `json:"statusMessage"`
}

// The Invoke method

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

        // Retrieve the requested Smart Contract function and arguments
        function, args := stub.GetFunctionAndParameters()
        // Route to the appropriate handler function to interact with the ledger
        if function == "recordNews" {
                return s.recordNews(stub, args)
        } else if function == "getNews" {
                return s.getNews(stub, args)
        } else if function == "getNewsHistory" {
                return s.getNewsHistory(stub, args)
        } else if function == "getOriginNewsHistory" {
                return s.getOriginNewsHistory(stub, args)
        } else if function == "recordComments" {
                return s.recordComments(stub, args)
        } else if function == "getComment" {
                return s.getComment(stub, args)
        } else if function == "getNewsComments" {
                return s.getNewsComments(stub, args)
        } else if function == "getUserComments" {
                return s.getUserComments(stub, args)
        } else if function == "getNewsCommentHistory" {
                return s.getNewsCommentHistory(stub, args)
        } else if function == "recordUsers" {
                return s.recordUsers(stub, args)
        } else if function == "getUser" {
                return s.getUser(stub, args)
        } else if function == "getAllUsers" {
                return s.getAllUsers(stub)
        } else if function == "getUserHistory" {
                return s.getUserHistory(stub, args)
        }

        return shim.Error("Invalid Smart Contract function name.")
}


// 新增修改新聞

func (s *SmartContract) recordNews(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        if len(args) != 7 {
                return shim.Error("Incorrect number of arguments. Expecting 7")
        }

        indexName := "news"
        // "news" +  Origin + key
        indexKey , _ := stub.CreateCompositeKey(indexName,[]string{args[6], args[0]})
        allAurthor := strings.Split(args[3], ",")

        var news = News{NewsId: indexKey, Title: args[1], ReleaseDatetime: args[2], Author: allAurthor, Newsurl: args[4], Content: args[5], Origin: args[6] }

        newsAsBytes, _ := json.Marshal(news)
        err := stub.PutState(indexKey, newsAsBytes)
        if err != nil {
                return shim.Error(fmt.Sprintf("Failed to record news catch: %s", args[0]))
        }

        return shim.Success(nil)
}

// 取得新聞資訊 輸入: key = origin+ key

func (sc *SmartContract) getNews(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        newsIdIsString , _ := stub.CreateCompositeKey("news",[]string{args[0], args[1]})

        newsIdAsByte, _ := stub.GetState(newsIdIsString)

        // 若沒取得資料，則回應系統錯誤
        if newsIdAsByte == nil {
                return shim.Error("Could not locate newsId")
        }

        // 若取得資料，將完整資料傳回
        return shim.Success(newsIdAsByte)
}

//取得新聞歷史資訊 key = origin + key

func (sc *SmartContract) getNewsHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        newsId , _ := stub.CreateCompositeKey("news",[]string{args[0], args[1]})

        // 取得該資產的歷史紀錄
        newsIter, _ := stub.GetHistoryForKey(newsId)
        defer newsIter.Close()

        // 歷史更改
        var newsHistory string
        for newsIter.HasNext() {
                result, _ := newsIter.Next()
                news := News{}

                // 將歷史value值 塞入news中
                json.Unmarshal(result.Value, &news)
                alterTime := result.Timestamp.String()
                author := strings.Join(news.Author," ")

                newsHistory +=  "title: " + news.Title + " , Content: " + news.Content+ " , Author: "+ author + " , url: "+ news.Newsurl +" , alterTime: " + alterTime+ "||"
        }
        // 將結果傳回
        return shim.Success([]byte(newsHistory))
}

//取得特定新聞台所有新聞資訊 輸入: origin

func (sc *SmartContract) getOriginNewsHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        indexName := "news"
        originName := args[0]

        resultsIterator,_ := stub.GetStateByPartialCompositeKey(indexName, []string{originName})
        defer resultsIterator.Close()

        var buffer bytes.Buffer
        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        fmt.Printf("- queryAllOriginNews:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}


// 新增修改評論

func (s *SmartContract) recordComments(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        if len(args) != 6 {
                return shim.Error("Incorrect number of arguments. Expecting 6")
        }

        indexName := "comment"

        // "comment"+  NewsId+ UserId+ key
        indexKey , _ := stub.CreateCompositeKey(indexName,[]string{args[4], args[5], args[0]})

        var comments = Comments{CommentId: indexKey, Comment: args[1], CommentDatetime: args[2], DisplayName: args[3], NewsId: args[4], UserId: args[5]}

        commentAsBytes, _ := json.Marshal(comments)
        err := stub.PutState(indexKey, commentAsBytes)
        if err != nil {
                return shim.Error(fmt.Sprintf("Failed to record comments catch: %s", args[0]))
        }

        return shim.Success(nil)
}

// 取得評論資訊 key = NewsId+ UserId+ key

func (sc *SmartContract) getComment(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        // 將用戶輸入轉換所需變數
        commentIdIsString , _ := stub.CreateCompositeKey("comment",[]string{args[0], args[1], args[2]})

        // 使用assetIdIsString向區塊鏈取得資料
        commentIdAsByte, _ := stub.GetState(commentIdIsString)

        // 若沒取得資料，則回應系統錯誤
        if commentIdAsByte == nil {
                return shim.Error("Could not locate commentIdAsByte")
        }

        // 若取得資料，將完整資料傳回
        return shim.Success(commentIdAsByte)
}

// 取得特定新聞評論 輸入: NewsId

func (sc *SmartContract) getNewsComments(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        indexName := "comment"
        newsId := args[0]

        resultsIterator,_ := stub.GetStateByPartialCompositeKey(indexName, []string{newsId})
        defer resultsIterator.Close()

        var buffer bytes.Buffer
        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        fmt.Printf("- queryAllNewsComments:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}

// 取得特定使用者評論 輸入: userId
/*
func (sc *SmartContract) getUserComments(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        userId := strings.ToLower(args[0])
        queryString := fmt.Sprintf("{\"selector\":{\"userId\":\"%s\"}}", userId)
        resultsIterator,err := stub.GetQueryResult(queryString)
        if err != nil {
                return shim.Error(err.Error())
        }
        defer resultsIterator.Close()

        var buffer bytes.Buffer
        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }
                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true

        }
        buffer.WriteString("]")

        fmt.Printf("- queryUsersComments:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}
*/

//取得評論歷史資訊 key = NewsId+ UserId+ key

func (sc *SmartContract) getNewsCommentHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        indexkey , _ := stub.CreateCompositeKey("comment",[]string{args[0], args[1], args[2]})

        // 取得該資產的歷史紀錄
        commentIter, _ := stub.GetHistoryForKey(indexkey)
        defer commentIter.Close()

        // 歷史更改
        var commentsHistory string
        for commentIter.HasNext() {
                result, _ := commentIter.Next()
                comments := Comments{}

                // 將歷史value值 塞入comments中
                json.Unmarshal(result.Value, &comments)
                alterTime := result.Timestamp.String()

                commentsHistory +=  "alterTime"+ alterTime+ "Comment" + comments.Comment +"||"
        }
        // 將結果傳回
        return shim.Success([]byte(commentsHistory))
}

// 新增修改使用者

func (s *SmartContract) recordUsers(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        if len(args) != 4 {
                return shim.Error("Incorrect number of arguments. Expecting 4")
        }

        indexName := "user"
        indexKey , _ := stub.CreateCompositeKey(indexName,[]string{"user", args[0]})

        var users = Users{UserId: indexKey, DisplayName: args[1], PictureUrl: args[2], StatusMessage: args[3]}

        usersAsBytes, _ := json.Marshal(users)
        err := stub.PutState(indexKey, usersAsBytes)
        if err != nil {
                return shim.Error(fmt.Sprintf("Failed to record users catch: %s", args[0]))
        }

        return shim.Success(nil)
}

// 取得使用者資訊 key = UserId

func (sc *SmartContract) getUser(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        userIdIsString , _ := stub.CreateCompositeKey("user",[]string{"user", args[0]})

        // 使用assetIdIsString向區塊鏈取得資料
        userIdAsByte, _ := stub.GetState(userIdIsString)

        // 若沒取得資料，則回應系統錯誤
        if userIdAsByte == nil {
                return shim.Error("Could not locate userIdAsByte")
        }

        // 若取得資料，將完整資料傳回
        return shim.Success(userIdAsByte)
}

// 取得所有使用者資訊

func (sc *SmartContract) getAllUsers(stub shim.ChaincodeStubInterface) sc.Response {

        indexName := "user"

        var buffer bytes.Buffer
        resultsIterator,_ := stub.GetStateByPartialCompositeKey(indexName, []string{"user"})
        defer resultsIterator.Close()

        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        fmt.Printf("- queryAllUsers:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}

// 取得使用者資料歷史修改資訊 key = newsId

func (sc *SmartContract) getUserHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

        indexkey , _ := stub.CreateCompositeKey("user",[]string{"user", args[0]})

        // 取得該資產的歷史紀錄
        userIter, _ := stub.GetHistoryForKey(indexkey)
        defer userIter.Close()

        // 歷史更改
        var userHistory string
        for userIter.HasNext() {
                result, _ := userIter.Next()
                user := Users{}

                // 將歷史value值 塞入user中
                json.Unmarshal(result.Value, &user)
                alterTime := result.Timestamp.String()

                userHistory +=  ", alterTime: "+ alterTime+ ", DisplayName: " + user.DisplayName+ ", PictureUrl: "+ user.PictureUrl+ ", StatusMessage: "+ user.StatusMessage+ "||"
        }
        // 將結果傳回
        return shim.Success([]byte(userHistory))
}



// 主程序
/*
 * main function *
calls the Start function
The main function starts the chaincode in the container during instantiation.
 */
func main() {

        // Create a new Smart Contract
        err := shim.Start(new(SmartContract))
        if err != nil {
                fmt.Printf("Error creating new Smart Contract: %s", err)
        }
}