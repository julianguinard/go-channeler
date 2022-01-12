package channeler

import (
    "time"
    "math/rand"
    "testing"
    "fmt"
    "bytes"
    "github.com/julianguinard/go-channeler/utils/strings"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cast"
)

var colors = []string{"red", "yellow", "green"}

type mapStringStringType map[string]string
type timeDurationByString map[string]int
type timeDurationByFruitAndColor map[string]timeDurationByString

/**
Usage : from inside docker container : /usr/local/go/bin/go test /go/src/app/utils/channeler/* -v
Test a Channeler's result and errors set when getRedApple callback returns error :
 must break getRedCherry which depends on it but every other result must be OK as there are no other callback
 in the chain which is depending on getRedApple
 */
func TestChanneler_RunRedAppleErrorPropagation(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 0,
            "red": 0,
            "green": 0,
        },
        "banana": timeDurationByString{
            "yellow": 0,
            "green": 0,
        },
        "cherry": timeDurationByString{
            "red": 0,
        },
    }
    channelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)

    //override standard getRedApple ChanneledCallback in order to return an error that must propagate
    //to other callbacks dependent on getRedApple (in this case, getRedCherry)
    redAppleErr := &(DependencyError{"getRedApple", "getRedCherry"})
    (*channelerInstance.CallbackChain)["getRedApple"] = NewChanneledCallback(
        func(dependencies CallbackResults) (interface{}, error) {
            return nil, redAppleErr
        }, []string{},
    )

    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    //assert getRedApple and getRedCherry have the same error propagated
    checkResultNilAndErrorsSet(channelerInstance, [][2]string{
        [2]string{"apple", "red"},
        [2]string{"cherry", "red"},
    }, redAppleErr, t)

    //assert other results are still ok
    checkResultsOkAndTimeAndNullError(channelerInstance, [][2]string{
        [2]string{"apple", "green"},
        [2]string{"apple", "yellow"},
        [2]string{"banana", "green"},
        [2]string{"banana", "yellow"},
    }, waitTimePerFruitAndColor, t)
    t.Logf("Channeler with error on getRedApple finished in %d seconds", roundedTimeInSeconds)
}

/**
Test a Channeler's result and errors set when getYellowApple callback returns error :
 must break getYellowBanana and getGreenBanana which depends on it but every other result must be OK as there are no other callback
 in the chain which is depending on getYellowApple
 */
func TestChanneler_RunYellowAppleErrorPropagation(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 0,
            "red": 0,
            "green": 0,
        },
        "banana": timeDurationByString{
            "yellow": 0,
            "green": 0,
        },
        "cherry": timeDurationByString{
            "red": 0,
        },
    }
    channelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)

    //now check that breaking getYellowApple dependency also breaks getYellowBanana and getGreenBanana
    yellowAppleErr := &DependencyError{"getYellowApple", "getYellowBanana and getGreenBanana"}
    (*channelerInstance.CallbackChain)["getYellowApple"] = NewChanneledCallback(
        func(dependencies CallbackResults) (interface{}, error) {
            return nil, yellowAppleErr
        }, []string{},
    )
    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    t.Logf("Channeler with error on getYellowApple finished in %d seconds", roundedTimeInSeconds)
    //assert getYellowApple, getYellowBanana and getGreenBanana have the same error propagated
    checkResultNilAndErrorsSet(channelerInstance, [][2]string{
        [2]string{"apple", "yellow"},
        [2]string{"banana", "green"},
        [2]string{"banana", "yellow"},
    }, yellowAppleErr, t)

    //assert other results are still ok
    checkResultsOkAndTimeAndNullError(channelerInstance, [][2]string{
        [2]string{"apple", "green"},
        [2]string{"apple", "red"},
        [2]string{"cherry", "red"},
    }, waitTimePerFruitAndColor, t)
}

/**
Test a Channeler's result and errors set when getYellowApple callback returns error :
 must break getYellowBanana and getGreenBanana which depends on it but every other result must be OK as there are no other callback
 in the chain which is depending on getYellowApple
 */
func TestChanneler_RunRedAppleAndYellowAppleErrorPropagation(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 0,
            "red": 0,
            "green": 0,
        },
        "banana": timeDurationByString{
            "yellow": 0,
            "green": 0,
        },
        "cherry": timeDurationByString{
            "red": 0,
        },
    }
    channelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)

    //override standard getRedApple ChanneledCallback in order to return an error that must propagate
    //to other callbacks dependent on getRedApple (in this case, getRedCherry)
    redAppleErr := &(DependencyError{"getRedApple", "getRedCherry"})
    (*channelerInstance.CallbackChain)["getRedApple"] = NewChanneledCallback(
        func(dependencies CallbackResults) (interface{}, error) {
            return nil, redAppleErr
        }, []string{},
    )
    yellowAppleErr := &DependencyError{"getYellowApple", "getYellowBanana and getGreenBanana"}
    (*channelerInstance.CallbackChain)["getYellowApple"] = NewChanneledCallback(
        func(dependencies CallbackResults) (interface{}, error) {
            return nil, yellowAppleErr
        }, []string{},
    )

    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    //assert getRedApple and getRedCherry have the same error propagated
    checkResultNilAndErrorsSet(channelerInstance, [][2]string{
        [2]string{"apple", "red"},
        [2]string{"cherry", "red"},
    }, redAppleErr, t)

    //assert getYellowApple, getYellowBanana and getGreenBanana have the same error propagated
    checkResultNilAndErrorsSet(channelerInstance, [][2]string{
        [2]string{"apple", "yellow"},
        [2]string{"banana", "green"},
        [2]string{"banana", "yellow"},
    }, yellowAppleErr, t)

    //the only ok result is getGreenBanana
    checkResultsOkAndTimeAndNullError(channelerInstance, [][2]string{
        [2]string{"apple", "green"},
    }, waitTimePerFruitAndColor, t)
    t.Logf("Channeler with error on getRedApple finished in %d seconds", roundedTimeInSeconds)
}

/**
Must execute in less than 1s as 0 waiting times are specified
 */
func TestChanneler_RunAllOkImediately(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 0,
            "red": 0,
            "green": 0,
        },
        "banana": timeDurationByString{
            "yellow": 0,
            "green": 0,
        },
        "cherry": timeDurationByString{
            "red": 0,
        },
    }

    channelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)
    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    assert.Equal(t, 0, roundedTimeInSeconds)
    //all results must be set
    checkResultsOkAndTimeAndNullError(channelerInstance, [][2]string{
        [2]string{"apple", "red"},
        [2]string{"apple", "yellow"},
        [2]string{"apple", "green"},
        [2]string{"banana", "yellow"},
        [2]string{"banana", "green"},
        [2]string{"cherry", "red"},
    }, waitTimePerFruitAndColor, t)
    t.Logf("This is the channeler's results obtained in %d seconds : %s...", roundedTimeInSeconds, channelerInstance.Results)
}

/**
Must execute in 11 seconds in optimal parallelisation status, knowing that we gather:
- 3 apples : red (1s), yellow (3s), green (6s)
- 2 bananas : yellow (4s) and green (5s), DEPENDENCIES : can be triggered only once we got green and yellow apples
- 1 cherry : red, DEPENDENCY : can be triggered only once we got red apple

Time result explanation

TIME IN SECONDS  0==========11
getRedApple      |1
getGreenApple    |=====6
getYellowApple   |==3
getYellowBanana        |===4
getGreenBanana         |====5
getRedCherry      |=====6

it takes 11 seconds ( longest times addition : 6 for getting getGreenApple + 5 for getting getGreenBanana after getGreenApple and getYellowApple are successfully retrieved)
in order to finish parallel treatments
 */
func TestChanneler_RunAllOkIn11s(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 3,
            "red": 1,
            "green": 6,
        },
        "banana": timeDurationByString{
            "yellow": 4,
            "green": 5,
        },
        "cherry": timeDurationByString{
            "red": 6,
        },
    }

    channelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)
    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    assert.Equal(t, 11, roundedTimeInSeconds)
    //all results must be set
    checkResultsOkAndTimeAndNullError(channelerInstance, [][2]string{
        [2]string{"apple", "red"},
        [2]string{"apple", "yellow"},
        [2]string{"apple", "green"},
        [2]string{"banana", "yellow"},
        [2]string{"banana", "green"},
        [2]string{"cherry", "red"},
    }, waitTimePerFruitAndColor, t)
    t.Logf("This is the channeler's results obtained in %d seconds : %s...", roundedTimeInSeconds, channelerInstance.Results)
}

/**
Must execute in 14 seconds in optimal parallelisation status, knowing that we gather:
1 - Fruits to make a jam (obtained in 11s like in the test above)
    - 3 apples : red (1s), yellow (3s), green (6s)
    - 2 bananas : yellow (4s) and green (5s), DEPENDENCIES : can be triggered only once we got green and yellow apples
    - 1 cherry : red, DEPENDENCY : can be triggered only once we got red apple
2 - A recipe book (obtained in 3s)
    - a recipe book to prepare apples (1s)
    - a recipe book to prepare bananas (2s)
    - a recipe book to prepare cherries (3s)
3 - A jam made from fruits and the recipe book : mix everything and return jam (3s) DEPENDENCIES : can be triggered only once we got fruits and recipe books


Time result explanation

TIME IN SECONDS         0=============14

===GET FRUITS===
getRedApple             |1
getGreenApple           |=====6
getYellowApple          |==3
getYellowBanana               |===4
getGreenBanana                |====5
getRedCherry             |=====6
===GET RECIPE BOOKS==
getAppleRb               |1
getBananaRb              |=2
getCherryRb              |==3
===GET JAM getJam====              |==3

it takes 11 seconds ( longest times addition : 6 for getting getGreenApple + 5 for getting getGreenBanana after getGreenApple and getYellowApple are successfully retrieved)
in order to finish parallel treatments
 */
func TestRecursiveChannelers_RunAllOkIn14s(t *testing.T) {
    var waitTimePerFruitAndColor = timeDurationByFruitAndColor{
        "apple": timeDurationByString{
            "yellow": 3,
            "red": 1,
            "green": 6,
        },
        "banana": timeDurationByString{
            "yellow": 4,
            "green": 5,
        },
        "cherry": timeDurationByString{
            "red": 6,
        },
    }
    var waitTimePerFruitBook = timeDurationByString{
        "apple": 1,
        "banana": 2,
        "cherry": 3,
    }
    var jamPreparationTime = 3

    channelerInstance := NewChanneler(&CallbackChain{
        "getFruits": NewChanneledCallback(
            func(dependencies CallbackResults) (interface{}, error) {
                fruitsChannelerInstance := initFruitsChannelerWithStandardCbChain(t, waitTimePerFruitAndColor)
                fruitsChannelerInstance.Run()
                return fruitsChannelerInstance.Results, nil
            }, []string{},
        ),
        "getRecipeBooks": NewChanneledCallback(
            func(dependencies CallbackResults) (interface{}, error) {
                recipeBooksChannelerInstance := NewChanneler(&CallbackChain{})
                for oneFruit, waitTime := range waitTimePerFruitBook {
                    oneFruit := oneFruit
                    (*recipeBooksChannelerInstance.CallbackChain)[oneFruit] = NewChanneledCallback(
                        func(dependencies CallbackResults) (interface{}, error) {
                            time.Sleep(time.Duration(waitTime) * time.Second)
                            return mapStringStringType{"book":oneFruit+" recipe book", "time": fmt.Sprintf("%d", waitTime)}, nil
                        }, []string{},
                    )
                }
                recipeBooksChannelerInstance.Run()
                return recipeBooksChannelerInstance.Results, nil
            }, []string{},
        ),
        "getJam": NewChanneledCallback(
            func(dependencies CallbackResults) (interface{}, error) {
                t.Logf("getJam started because we received %s", dependencies)
                time.Sleep(time.Duration(jamPreparationTime) * time.Second)
                var buffer bytes.Buffer
                buffer.WriteString("This is a delicious jam made of these fruits : ")
                for _, oneResultOfGetFruits := range dependencies["getFruits"].(CallbackResults) {
                    buffer.WriteString(" "+oneResultOfGetFruits.(mapStringStringType)["fruit"]+",")
                }
                buffer.WriteString(", using these recipe books : ")
                for _, oneResultOfGetBooks := range dependencies["getRecipeBooks"].(CallbackResults) {
                    buffer.WriteString(" "+oneResultOfGetBooks.(mapStringStringType)["book"]+",")
                }

                return buffer.String(), nil
            }, []string{"getFruits", "getRecipeBooks"}),
    })
    start := time.Now()
    channelerInstance.Run()
    elapsed := time.Since(start)
    roundedTimeInSeconds := int(elapsed / time.Second)
    assert.Equal(t, 14, roundedTimeInSeconds)
}

func getFruit(t *testing.T, fruitName string, color string, waitTimePerFruitAndColor timeDurationByFruitAndColor) mapStringStringType {
    t.Logf("start getting %s %s...", color, fruitName)
    waitTime, isset := waitTimePerFruitAndColor[fruitName][color]
    if !isset {
        waitTime = getRandomWaitTime()
    }
    time.Sleep(time.Duration(waitTime) * time.Second)
    t.Logf("get%s : Got %s %s after %d seconds", fruitName, color, fruitName, waitTime)
    return mapStringStringType{
        "fruit":fruitName + " " + color,
        "time": fmt.Sprintf("%d", waitTime),
    }
}

/**
Instanciate a channeler to get the following fruit set
1 - 3 apples (red, yellow, green)
2 - 2 bananas : yellow and green, DEPENDENCIES : can be triggered only once we got green and yellow apples
3 - 1 cherry : red, DEPENDENCY : can be triggered only once we got red apple
Callback names are named after the folowwing form : "get<Color><Fruitname>"
 */
func initFruitsChannelerWithStandardCbChain(t *testing.T, waitTimePerFruitAndColor timeDurationByFruitAndColor) *Channeler {
    callbackChain := CallbackChain{}
    channelerInstance := NewChanneler(&callbackChain)

    for _, oneColor := range colors {
        oneColor := oneColor
        callbackChain[fmt.Sprintf("get%sApple", strings.Ucfirst(oneColor))] = NewChanneledCallback(
            func(dependencies CallbackResults) (interface{}, error) {
                return getFruit(t, "apple", oneColor, waitTimePerFruitAndColor), nil
            }, []string{},
        )
    }
    //banana needs green and yellow apple dependency before being able to run
    for _, oneColor := range colors[1:] {
        oneColor := oneColor
        callbackName := fmt.Sprintf("get%sBanana", strings.Ucfirst(oneColor))
        callbackChain[callbackName] = NewChanneledCallback(
            //it is expected to get the yellow apple as 1st and only arg
            func(dependencies CallbackResults) (interface{}, error) {
                t.Logf("%s started because we received %s", callbackName, dependencies)
                return getFruit(t, "banana", oneColor, waitTimePerFruitAndColor), nil
            }, []string{"getGreenApple", "getYellowApple"},
        )
    }
    //cherry needs red apple dependency before being able to run
    callbackChain[fmt.Sprintf("get%sCherry", strings.Ucfirst(colors[0]))] = NewChanneledCallback(
        func(dependencies CallbackResults) (interface{}, error) {
            t.Logf("getRedCherry started because we received %s", dependencies)
            return getFruit(t, "cherry", colors[0], waitTimePerFruitAndColor), nil
        }, []string{"getRedApple"},
    )
    return channelerInstance
}

/**
return a random integer between 1 and 20, to be casted as a "Second" time.Duration
 */
func getRandomWaitTime() int {
    rand.Seed(time.Now().UTC().UnixNano())
    returnedSecondsNb := rand.Intn(20) + 1
    return returnedSecondsNb
}

func checkResultsOkAndTimeAndNullError(channelerInstance *Channeler, fruitAndColorSlices [][2]string, waitTimePerFruitAndColor timeDurationByFruitAndColor, t *testing.T) {
    //assert other results are still ok
    for _, oneFruitColorCouple := range fruitAndColorSlices {
        cbName := fmt.Sprintf("get%s%s", strings.Ucfirst(oneFruitColorCouple[1]), strings.Ucfirst(oneFruitColorCouple[0]))
        assert.Equal(t, oneFruitColorCouple[0]+" "+oneFruitColorCouple[1], channelerInstance.Results[cbName].(mapStringStringType)["fruit"])
        assert.Equal(t, waitTimePerFruitAndColor[oneFruitColorCouple[0]][oneFruitColorCouple[1]], cast.ToInt(channelerInstance.Results[cbName].(mapStringStringType)["time"]))
        assert.Equal(t, nil, channelerInstance.Errors[cbName])
    }
}

func checkResultNilAndErrorsSet(channelerInstance *Channeler, fruitAndColorSlices [][2]string, searchedError error, t *testing.T) {
    for _, oneFruitColorCouple := range fruitAndColorSlices {
        cbName := fmt.Sprintf("get%s%s", strings.Ucfirst(oneFruitColorCouple[1]), strings.Ucfirst(oneFruitColorCouple[0]))
        assert.Equal(t, nil, channelerInstance.Results[cbName])
        assert.Equal(t, searchedError, channelerInstance.Errors[cbName])
    }
}