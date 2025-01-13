//go:build scheduler.cores

#include <inttypes.h>

__attribute__((naked))
void tinygo_cores_startTask(void) {
    asm volatile(
        "bl tinygo_schedulerUnlock\n\t"

        "ldr r0, =tinygo_exitTask\n\t"
        "mov lr, r0\n\t"

        "pop {r0, pc}\n\t"
    );
}

void tinygo_switchTask(uintptr_t *oldStack, uintptr_t newStack) {
#if defined(__thumb__)
    register uintptr_t *oldStackReg asm("r0");
    oldStackReg = oldStack;
    register uintptr_t newStackReg asm("r1");
    newStackReg = newStack;
    asm volatile(
        // Push PC to switch back to.
        // Note: adding 1 to set the Thumb bit.
        "ldr r2, =1f+1\n\t"
        "push {r2}\n\t"

        // Save stack pointer in oldStack for the switch back.
        "mov r2, sp\n\t"
        "str r2, [%[oldStack]]\n\t"

        // Switch to the new stack.
        "mov sp, %[newStack]\n\t"

        // Return into the new stack.
        "pop {pc}\n\t"

        // address where we should resume
        "1:"

        : [oldStack]"+r"(oldStackReg),
          [newStack]"+r"(newStackReg)
        :
        : "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r11", "r12", "lr", "cc", "memory"
    );
#else
    #error unknown architecture
#endif
}
