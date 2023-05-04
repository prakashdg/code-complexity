package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"flag"
	"github.com/olekukonko/tablewriter"
	"strconv"
	"regexp"
)

func changeset_generator(list_of_fun *[]string, changeset_file string) {

	// Function to generate list of files changed and their respective function names
	data, _ := ioutil.ReadFile(changeset_file)
	contents := strings.Split(string(data), "\n")

	for iter := 0; iter < len(contents); iter++ {
		function_name := strings.Split(contents[iter], " @@ ")
		if len(function_name) > 1 {
			*list_of_fun = append(*list_of_fun, function_name[1])
		}
	}
}

func analyze_complexity_report(complexity_data string) map[string]map[string]string {

	//Function to generate complexity based on changed functions list
	data, _ := ioutil.ReadFile(complexity_data)
	var complex_map = map[string]map[string]string{}
	contents := strings.Split(string(data), "\n")
	for iter := 0; iter < len(contents); iter++ {
		list_of_elem := strings.Fields(contents[iter])
		if len(list_of_elem) == 5 {
			complex_map[list_of_elem[4]] = map[string]string{}
			complex_map[list_of_elem[4]]["filename"] = list_of_elem[3]
			complex_map[list_of_elem[4]]["score"] = list_of_elem[0]
			complex_map[list_of_elem[4]]["ln-c"] = list_of_elem[1]
			complex_map[list_of_elem[4]]["nc-lns"] = list_of_elem[2]
		}
	}
	return complex_map
}

func find_new_func_name(source map[string]map[string]string, dest map[string]map[string]string) []string {	
	var new_func []string
	for key, _:=range source {
		if _, value:=dest[key]; !value {
			new_func=append(new_func, key)
		}
	}
	return new_func
}

func check_mathcing_func(func_list []string, cond_name string) bool {
	for _, elem := range func_list {

		if strings.Contains(elem, cond_name) {
			return true
		}
	}
	return false
}

func generate_complexity_match(complex_map map[string]map[string]string, changest_list []string) []string {
	var new_list []string
	for key, _ := range complex_map {
		if check_mathcing_func(changest_list, key) {
			new_list = append(new_list, key)
		}
	}
	return new_list
}

func generate_markdown_table(base map[string]map[string]string, source map[string]map[string]string, funcs []string) string {

	// Generates markdown table for complexity analysis
	var filename,func_name,base_lc, base_score,source_lc,source_score string
	var base_nc_lns, color string
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"FileName", "FuncName", "T-fs", "S-fs", "T-lpf", "S-lpf",})
	table.Append([]string{"--------", "--------", "--------", "--------", "--------", "--------"})
	for _, elem := range funcs {
		if _, cont:=base[elem]; cont {
			base_nc_lns=base[elem]["nc-lns"]
			base_score=base[elem]["score"]
			color="red"
		} else {
			base_nc_lns="nil"
			base_score="0"
			color="blue"
		}
		source_score_value,_:=strconv.Atoi(source[elem]["score"])
		target_score,_:=strconv.Atoi(base_score)
		if source_score_value > target_score {
			filename=source[elem]["filename"]
			update_func:=regexp.MustCompile(`_`)
			elem_new:=update_func.ReplaceAllString(elem, "\\_")
			func_name=fmt.Sprintf("$`\\textcolor{%s}{\\text{*%s}}`$", color, elem_new )
			base_lc=fmt.Sprintf("$`\\textcolor{%s}{\\text{%s}}`$", color, base_nc_lns)
			base_score=fmt.Sprintf("$`\\textcolor{%s}{\\text{%s}}`$", color, base_score)
			source_lc=fmt.Sprintf("$`\\textcolor{%s}{\\text{%s}}`$", color, source[elem]["nc-lns"])
			source_score=fmt.Sprintf("$`\\textcolor{%s}{\\text{%s}}`$", color, source[elem]["score"])
		} else {
			filename=source[elem]["filename"]
			func_name=elem
			base_lc=base_nc_lns
			base_score=base_score
			source_lc=source[elem]["nc-lns"]
			source_score=source[elem]["score"]			
		}
		table.Append([]string{filename, func_name, base_score, source_score, base_lc, source_lc})
	}
	// Render the table as markdown
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetHeaderLine(false)
	table.SetAutoWrapText(false)
	table.Render()
	return tableString.String()
}

func main() {
	var sourceComplexity string
	var targetComplexity string
	var mrID string
	var changesetFile string

	flag.StringVar(&sourceComplexity, "sourceComplexity", "", "Supply source branch complexity Values")
	flag.StringVar(&targetComplexity, "targetComplexity", "", "Supply target branch complexity Values")
	flag.StringVar(&changesetFile, "changesetFile", "", "Specify the file name where git diff saved")
	flag.StringVar(&mrID, "mrID", "", "Supply merge request ID")

    flag.Parse()

	var list_of_fun []string
	changeset_generator(&list_of_fun, changesetFile)
	complex_dict_source := analyze_complexity_report(sourceComplexity)
	//fmt.Printf("Element dict is %s", complex_dict_source)
	matched_func := generate_complexity_match(complex_dict_source, list_of_fun)
	//for _, elem := range matched_func {
	//	fmt.Printf("Matched Functions are %s\n", elem)
	//}
	complex_dict_base := analyze_complexity_report(targetComplexity)
	new_func:=find_new_func_name(complex_dict_source,complex_dict_base)
	matched_func=append(matched_func, new_func...)
	markdown_values := generate_markdown_table(complex_dict_base, complex_dict_source, matched_func)
	fmt.Printf("values are %s", markdown_values)
	result := get_mr_desc(mrID)
	fmt.Printf("\n ****** MR Current description is **** \n %s", result)
	update_mr(mrID, markdown_values, result)
}
