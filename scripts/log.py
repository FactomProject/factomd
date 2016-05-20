import time
import string
import sys
import commands

def get_cpumem(pid):
    d = [i for i in commands.getoutput("ps aux").split("\n")
        if i.split()[1] == str(pid)]
    return (float(d[0].split()[2]), float(d[0].split()[3]) ) if d else None

if __name__ == '__main__':
    if not len(sys.argv) == 2 or not all(i in string.digits for i in sys.argv[1]):
        print("usage: %s PID" % sys.argv[0])
        sys.stdout.flush()
        exit(2)
    print("%CPU\t%MEM")
    sys.stdout.flush()
    try:
        while True:
            x = get_cpumem(sys.argv[1])
            if not x:
                print("no such process")
                sys.stdout.flush()
                exit(1)
            print("%.2f\t%.2f\t" % x)
            sys.stdout.flush()
            time.sleep(10)
    except KeyboardInterrupt:
        print
        exit(0)
