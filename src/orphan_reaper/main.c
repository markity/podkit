#include <unistd.h>
#include <sys/wait.h>
 #include <time.h>

int main() {
    while(1) {
        int status;
        int pid = waitpid(-1, &status, 0);
        if (pid == -1) {
            sleep(1);
        }
    }
}
