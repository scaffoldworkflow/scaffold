command = ""
while True:
    line = input()
    while line != "marathon::script_end":
        command += f'{line}\n'
        line = input()
    try:
        exec(command, globals(), locals())
    except Exception as e:
        print(e)
        print('marathon::exception')
    command = ""
