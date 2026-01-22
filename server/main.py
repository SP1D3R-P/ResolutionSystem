

from dotenv import load_dotenv 

from multiprocessing import Process
import os 

import pathlib
from pathlib import Path

def run_go_server():
    go_file : Path = Path(__file__).parent / 'GORPC/serve.go'
    os.system(f"go run {go_file}")


if __name__ == "__main__":
    load_dotenv()

    # This needed to be here 
    from PyRPC import serve

    # p : list[Process] = []

    # p.append(Process(target=run_go_server))
    # p.append(Process(target=serve.serve))
    

    # for proc in p : 
    #     proc.start()

    # for proc in p : 
    #     proc.join()


    serve.serve()