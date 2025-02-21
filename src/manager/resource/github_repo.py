import sys
import json
import subprocess

def do_in(args):
    # Args should have the following keys:
    # - url
    # - branch
    subprocess.run(["git", "clone", "-b", args['branch'], args['url']])

def do_out(args):
    print('git repo out not yet implemented')

if __name__ == '__main__':
    direction = sys.argv[1] # can be either 'in' or 'out'
    args = json.loads(sys.argv[2]) # JSON of arguments for the input/output
    if direction == 'in':
        do_in(args)
    elif direction == 'out':
        do_out(args)
    else:
        print(f"Invalid direction: {direction}, allowed directions are 'in' and 'out'")
