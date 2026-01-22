
import os
import dotenv
import sys 
import pathlib

if __name__ == "__main__" :
    dotenv.load_dotenv()
    currdir = pathlib.Path(__file__).parent
    if len(sys.argv) != 2 : 
        raise ValueError("Usage : uv run main.py <dev/user>")
        

    match sys.argv[1] :
        case 'dev' :
            os.system(f'go run {currdir / "dev/main.go"}')
        case 'user' :
            os.system(f"go run {(currdir / 'user/main.go').relative_to(currdir.parent)}")
        case _: 
            raise ValueError("Usage : uv run main.py <dev/user>")
            


