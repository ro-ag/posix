#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <libgen.h>
#include <sys/mman.h>

typedef struct {
    union {
        void *pVoid;
        uintptr_t uintptr;
    };
    size_t memSize;
} address_t;

typedef struct {
    address_t address;
    size_t textSize;
    char *textPtr;
} handler_t;

const int FileDesc = 3;

int main(int argc, char *argv[]) {
    printf("(%s) C Program\n", basename(argv[0]));
    address_t *pAdd, add;
    /* Get start address */
    if ((pAdd = mmap(0, sizeof(address_t), PROT_READ, MAP_SHARED, FileDesc, 0)) == MAP_FAILED) {
        perror("unable to map file");
        exit(EXIT_FAILURE);
    }

    add = *pAdd;
    if (munmap(pAdd, sizeof(address_t)) == -1) {
        perror("unable to unmap address");
        exit(EXIT_FAILURE);
    }

    /* mmap again with fixed address */

    handler_t *h = NULL;
    if ((h = mmap(add.pVoid, add.memSize, PROT_READ, MAP_SHARED | MAP_FIXED, FileDesc, 0)) == MAP_FAILED) {
        perror("unable to map fixed");
        exit(EXIT_FAILURE);
    }

    printf("(%s) Got Text: %.*s\n", basename(argv[0]), (int) h->textSize, h->textPtr);

    if (munmap(h, add.memSize) == -1) {
        perror("unable to unmap handler");
        exit(EXIT_FAILURE);
    }
    return 0;
}