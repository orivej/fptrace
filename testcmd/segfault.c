int main() {
    char *ptr = (char *)100;
    for (;;) {
        ++*ptr;
        ++ptr;
    }
}
