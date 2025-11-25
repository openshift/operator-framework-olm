import os
import sys

if sys.version_info < (3, 9):
    sys.exit("Sorry, Python < 3.9 is no longer supported.")

sys.dont_write_bytecode = True

def version():
    return "0.1"
