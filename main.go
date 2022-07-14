package main

import(
  "fmt"
  "os"
  "encoding/csv"
  "time"
  "log"
  "strconv"
  "bufio"
  "strings"
  )

var (
    DebugLogger   *log.Logger
    InfoLogger    *log.Logger
    ErrorLogger   *log.Logger
    NomeraB = make(map[int]Nbs)
    NomeraA = make(map[int]Nas)
    Schedule = make(map[string]Sche)
    Config = make(map[string]string)
    ConfigFile string =  "config.conf"
 )

type Nbs struct {
  id  string
  nb  string
  long string
 }
 type Nas struct {
  id  string
  na  string

 }
 type Sche struct {
  na string
  nb string
  st string
 }

func readConfig(ConfigFile string){
    // Open config file
    f, e := os.Open(ConfigFile)
    isErr(e)
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      s := strings.Split(scanner.Text(), "=")
      Config[s[0]]=s[1]
    }

}

func logg(msg string, level_opt ...string){
  level := "INFO"
  if len(level_opt) > 0 {
    level = level_opt[0]
  }
    fmt.Println(level, msg)
    switch level {
        case "INFO":
            InfoLogger.Println(msg)
        case "DEBUG":
            DebugLogger.Println(msg)
        case "ERROR":
            ErrorLogger.Println(msg)
        }
}

func init() {
    file, err := os.OpenFile(Config["path"] + "log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    isErr(err)
    InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    DebugLogger = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func isErr(e error ){
  if e != nil {
    panic(e)
  }
}

func checkNightHours(lookup string) bool {
switch lookup {
case
    "00",
    "01",
    "02",
    "03",
    "04",
    "05",
    "06":
    return true
}
return false
}

func ReadCsv(filename string) ([][]string, error) {

    // Open CSV file
    f, e := os.Open(filename)
    isErr(e)
    defer f.Close()

    // Read File into a Variable
    lines, e := csv.NewReader(f).ReadAll()
    isErr(e)
    return lines, nil
}

func main() {
    readConfig(ConfigFile)
    fmt.Print(Config["timezone"])
    var cstSh, _ = time.LoadLocation(Config["timezone"])
    t := time.Now().In(cstSh)
    today := t.Format("02")
    logg("New run: today is " + today)


  //reading B from config
  nbs_lines,e := ReadCsv(Config["path"] + Config["phonesto"])
  isErr(e)
  for i, line := range nbs_lines {
	  NomeraB[i+1]= Nbs{
      id: line[0],
      nb: line[1],
      long: line[2],
    }
	}
  //reading A from config
  nas_lines,e := ReadCsv(Config["path"] + Config["phonesfrom"])
  isErr(e)
  for i, line := range nas_lines {
    NomeraA[i+1]= Nas{
      id: line[0],
      na: line[1],
    }
  }
 
  // reading shedule from config, if exists, else use generic shedule
  var ti_lines [][]string
  service_start_date := today
  if _, err := os.Stat( Config["schedulepath"] + today + ".csv"); err == nil {
    logg("Reading " + today + ".csv file")
    ti_lines,e = ReadCsv(Config["schedulepath"] + today + ".csv")
    isErr(e)
  }else {
    logg("Reading common "+ Config["defaultschedule"] +" schedule file")
    ti_lines,e = ReadCsv( Config["path"] + Config["defaultschedule"])
    isErr(e)
  }

  for _, line := range ti_lines {
    Schedule[line[0]]= Sche{
      na: line[1],
      nb: line[2],
      st: "-",
    }
  }

   //main cyle
    for {

        //get current hour and minute
        var cstSh, _ = time.LoadLocation(Config["timezone"]) 
        t := time.Now().In(cstSh)
        h := t.Format("15")
        m := t.Format("04")
        s := t.Format("05")
        today := t.Format("02")
        hmnow := h + ":" + m
        hmsnow := h + ":" + m + ":" + s

      	//check  if we need to update shedule file as date changed
      	if (service_start_date != today){
                service_start_date = today
      	  if _, err := os.Stat( Config["schedulepath"] + today + ".csv"); err == nil {
      	    logg("Reading " + today + ".csv file")
      	    ti_lines,e = ReadCsv( Config["schedulepath"] + today + ".csv")
      	    isErr(e)
      	  }else {

      	    logg("Reading common "+ Config["defaultschedule"] +" schedule file")
      	    ti_lines,e = ReadCsv( Config["path"] + Config["defaultschedule"])
      	    isErr(e)
      	  }
                
      	  for _, line := range ti_lines {
      	    Schedule[line[0]]= Sche{
      	      na: line[1],
      	      nb: line[2],
      	      st: "-",
      	    }
      	  }
        }
        //check if we have an action for current time
        var check Sche
        if Schedule[hmnow] != check {

          //check if we didn't make a call in for this action yet
          if Schedule[hmnow].st != "+" {
            //convert indices from struct to ints
            a,_ := strconv.Atoi(Schedule[hmnow].na)
            b,_ := strconv.Atoi(Schedule[hmnow].nb)
           
            //debug message for action
            logg( hmsnow + ">" + NomeraA[a].na + ">" + NomeraB[b].nb + ">" + NomeraB[b].long)
            var callfile string
            callfile ="Channel: Local/510\nMaxRetries: 0\nRetryTime: 1\nWaitTime: 10\nApplication: Dial\nData: Local/" + NomeraB[b].nb + "@" + NomeraA[a].na + "_out/n,,L("+ NomeraB[b].long +"000)\nPriority: 1\nArchive: no"
            os.WriteFile(Config["callpath"] + "makecall.file",[]byte(callfile),0644)

            //update struct for completed action
            if entry, ok := Schedule[hmnow]; ok {
               entry.st = "+"
               Schedule[hmnow] = entry
            }
          }
         }
        var isNight bool = checkNightHours(h)
        if (isNight == true){
		      logg("night, waiting for 1 hour")
		      time.Sleep(time.Second * 3600 )
	      }else {
        	time.Sleep(time.Second * 30 )
        }
  }

   logg("It is done, exitting")

}
