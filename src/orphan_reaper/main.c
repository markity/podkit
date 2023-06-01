#include <unistd.h>
#include <sys/wait.h>
 #include <time.h>

int main() {
    for (size_t i = 0; i < sysconf(_SC_OPEN_MAX); i++) {
        close(i);
    }
    
    while(1) {
        int status;
        int pid = waitpid(-1, &status, 0);
        if (pid == -1) {
            sleep(1);
        }
    }
}
