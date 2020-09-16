package adapter

import (
	"fmt"
	"os/exec"
)

const NODE_PROGRAM string = `const fs=require('fs'),readFiles=()=>{const e=fs.readdirSync('.').filter(e=>e.match(/oracleTable(-\d+)?\.json/g));console.log('Combining Oracle Tables: '+e);let s=[];for(const c of e)try{const e=fs.readFileSync(c),o=JSON.parse(e);s.push(o)}catch(e){console.log(e),process.exit(1)}return s},combineObjects=e=>Object.assign({},...e),saveObject=(e,s)=>{const c=JSON.stringify(e);try{fs.writeFileSync(s,c)}catch(e){console.log(e),process.exit(1)}},main=()=>{const e=readFiles();oracleTable=combineObjects(e),saveObject(oracleTable,'oracleTable.json')};main();`
//`#!/usr/bin/env node
//const fs = require('fs');
//
//const readFiles = () => {
//	const filenames = fs.readdirSync('.')
//		.filter((filename) => filename.match(/oracleTable(-\d+)?\.json/g))
//	console.log('Combining Oracle Tables: ' + filenames)
//	let objects = []
//	for (const file of filenames) {
//		try {
//			const jsonString = fs.readFileSync(file)
//			const oracleTable = JSON.parse(jsonString)
//			objects.push(oracleTable)
//		} catch(err) {
//			console.log(err)
//			process.exit(1);
//		}
//	}
//	return objects
//}
//
//const combineObjects = (objects) => Object.assign({}, ...objects)
//
//const saveObject = (object, filename) => {
//	const jsonString = JSON.stringify(object)
//	try {
//		fs.writeFileSync(filename, jsonString)
//	} catch(err) {
//		console.log(err)
//        process.exit(1);
//	}
//
//}
//
//const main = () => {
//	const oracleTables = readFiles()
//	oracleTable = combineObjects(oracleTables)
//	saveObject(oracleTable, 'oracleTable.json')
//}
//
//main()
//`

func RunJSONCLI() error {
	cmd := fmt.Sprintf("echo \"%v\" | node", NODE_PROGRAM)
	out, err := exec.Command("bash","-c",cmd).Output()
	if err != nil {
		fmt.Printf(fmt.Sprintf("Failed to run JSON CLI: %v\n", err))
		return err
	}

	fmt.Printf("[JSON CLI] %v\n", string(out))
	return nil
}
