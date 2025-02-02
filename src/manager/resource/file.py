import sys
import json

def do_in(args):
    with open(args['path'], 'w', encoding='utf-8') as out_file:
        out_file.write('Hello world!')

def do_out(args):
    with open(args['path'], 'r', encoding='utf-8') as in_file:
        print(in_file.read())

if __name__ == '__main__':
    direction = sys.argv[1] # can be either 'in' or 'out'
    args = json.loads(sys.argv[2]) # JSON of arguments for the input/output
    if direction == 'in':
        do_in(args)
    elif direction == 'out':
        do_out(args)
    else:
        print(f"Invalid direction: {direction}, allowed directions are 'in' and 'out'")
