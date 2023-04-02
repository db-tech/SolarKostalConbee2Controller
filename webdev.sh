#!/bin/sh


# check if npm is installed
if ! command -v npm &> /dev/null
then
  # print error message if npm is not installed
  echo "Error: npm is not installed. Please install npm and try again."
  exit 1
fi

# continue with the rest of your script here

# check if yarn is installed
if ! command -v yarn &> /dev/null
then
  # print error message if yarn is not installed
  echo "Error: yarn is not installed. Please install yarn and try again."
  exit 1
fi

cd web || exit

# check if the node_modules directory exists
if [ ! -d "node_modules" ]
then
  # if node_modules does not exist, run yarn to download dependencies
  yarn
fi

cd - || exit


go install github.com/cosmtrek/air@latest
air && fg & cd web && yarn devwatch