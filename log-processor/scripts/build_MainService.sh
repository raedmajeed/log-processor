AppRoot='../'

function moduleInit()
{
  sh ./moduleInit.sh
}

function make_all() 
{
  echo "Building the project"
	cd $AppRoot/build
	make -f Makefile
	cd -
}

function make_docker()
{
  echo "Creating Docker Image"
  cd $AppRoot/build
  make -f Makefile docker
	cd -
}

if [[ "$1" == "make_all" ]]; then
    moduleInit
    make_all
elif [[ "$1" == "make_docker" ]]; then
    moduleInit
    make_docker
else
    echo "select some parameter"
fi