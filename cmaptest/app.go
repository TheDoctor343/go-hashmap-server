package main
import "fmt"

//const mapSize int = 109
//const threads int = 109
//const mapSize int = 3
//const threads int = 3
const mapSize int = 19
const threads int = 19
//const mapSize int = 353
//const threads int = 353
func main() {
	//ChannelExample()
	//mapExample()

	/*Compares the time to read and write 2 million values*/
	fmt.Println("Time to read/write 2 million values:")

	series := mapTestDefault()

	//mapTestSeries(mapSize)

	concurrent := mapTestConcurrent(mapSize, threads)

	fmt.Printf("Concurrent was %.0f%% faster\n",float64(series-concurrent)/float64(series)*100)

	//mapTestSeries2(mapSize)

	concurrent = mapTestConcurrent2(mapSize, threads)

	fmt.Printf("V2: Concurrent was %.0f%% faster\n",float64(series-concurrent)/float64(series)*100)
}
