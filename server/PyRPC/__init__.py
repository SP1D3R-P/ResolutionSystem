
import os 
import sys 
import pathlib

# this is due to we are making module out of proto [ no relative import is there while genrating ]
sys.path.append(os.path.abspath(pathlib.Path(__file__).parent / 'proto' ))

from . import proto
from . import services