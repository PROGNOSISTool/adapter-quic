package adapter

import (
	"fmt"
	"os/exec"
	"syscall"
)

const NODE_PROGRAM string = `const fs=require('fs'),selectFiles=()=>fs.readdirSync('.').filter(e=>e.match(/oracleTable(-\d+)?\.json/g)),readFiles=e=>{console.log('Combining Oracle Tables: '+e);let s=[];for(const c of e)try{const e=fs.readFileSync(c),o=JSON.parse(e);s.push(o)}catch(e){console.log(e),process.exit(1)}return s},removeFiles=e=>{for(const s of e)fs.unlinkSync(s)},combineObjects=e=>Object.assign({},...e),saveObject=(e,s)=>{const c=JSON.stringify(e);try{fs.writeFileSync(s,c)}catch(e){console.log(e),process.exit(1)}},main=()=>{const e=selectFiles(),s=readFiles(e),c=combineObjects(s);saveObject(c,'oracleTable.json');const o=e.filter(e=>'oracleTable.json'!==e);removeFiles(o)};main();`
//#!/usr/bin/env node
//const fs = require('fs');
//
//const selectFiles = () => fs.readdirSync('.')
//.filter((filename) => filename.match(/oracleTable(-\d+)?\.json/g))
//
//const readFiles = (filenames) => {
//console.log('Combining Oracle Tables: ' + filenames)
//let objects = []
//for (const file of filenames) {
//try {
//const jsonString = fs.readFileSync(file)
//const oracleTable = JSON.parse(jsonString)
//objects.push(oracleTable)
//} catch(err) {
//console.log(err)
//process.exit(1);
//}
//}
//return objects
//}
//
//const removeFiles = (filenames) => {
//for (const file of filenames) {
//fs.unlinkSync(file)
//}
//}
//
//const combineObjects = (objects) => Object.assign({}, ...objects)
//
//const saveObject = (object, filename) => {
//const jsonString = JSON.stringify(object)
//try {
//fs.writeFileSync(filename, jsonString)
//} catch(err) {
//console.log(err)
//process.exit(1);
//}
//
//}
//
//const main = () => {
//const filesToMerge = selectFiles()
//const oracleTables = readFiles(filesToMerge)
//const oracleTable = combineObjects(oracleTables)
//saveObject(oracleTable, 'oracleTable.json')
//const filesToDelete = filesToMerge.filter((filename) => filename !== "oracleTable.json")
//removeFiles(filesToDelete)
//}
//
//main()

func hasNodeInstalled() (bool, error) {
    _, err := exec.Command("sh","-c","command -v node").Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					return false, nil
				}
			}
		} else {
			fmt.Printf(fmt.Sprintf("Failed to check for Node: %v\n", err))
			return false, err
		}
	}
	return true, nil
}

func RunJSONCLI() error {
	hasNode, err := hasNodeInstalled()
	if err != nil {
		return err
	}

	if !hasNode {
		err := fmt.Errorf("NodeJS not found.")
		fmt.Printf(fmt.Sprintf("Failed to run JSON CLI: %v\n", err))
		return err
	}

    cmd := fmt.Sprintf("echo \"%v\" | /usr/bin/env node", NODE_PROGRAM)
	out, err := exec.Command("sh","-c",cmd).Output()
	if err != nil {
		fmt.Printf(fmt.Sprintf("Failed to run JSON CLI: %v\n", err))
		return err
	}

	fmt.Printf("[JSON CLI] %v\n", string(out))
	return nil
}
