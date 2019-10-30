#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char s[5000];
int m[20000] = {32};
int length = 1;
int pc;
int dstack[500];
int *dsp = dstack;
int top = 64;
int w;
int st0;

void
at(int x)
{
	m[m[0]++] = length;
	length = *m - 1;
	m[m[0]++] = top;
	m[m[0]++] = x;
	scanf("%s", s + top);
	top += strlen(s + top) + 1;
}

void
run(int x)
{
	switch (m[x++]) {
	case 0: // pushint
		*++dsp = st0;
		st0 = m[pc++];
		break;
	case 1: // compile me
		m[m[0]++] = x;
		break;
	case 2: // run me
		m[++m[1]] = pc;
		pc = x;
		break;
	case 3: // :
		at(1);
		m[m[0]++] = 2;
		break;
	case 4: // immediate
		*m -= 2;
		m[m[0]++] = 2;
		break;
	case 5: // _read
		for (w = scanf("%s", s) < 1 ? exit(0), 0 : length; strcmp(s, &s[m[w + 1]]); w = m[w]);
		if (w - 1) {
			run(w + 2);
		} else {
			m[m[0]++] = 2;
			m[m[0]++] = atoi(s);
		}
		break;
	case 6: // @
		st0 = m[st0];
		break;
	case 7: // !
		m[st0] = *dsp--;
		st0 = *dsp--;
		break;
	case 8: // -
		st0 = *dsp-- - st0;
		break;
	case 9: // *
		st0 *= *dsp--;
		break;
	case 10: // /
		st0 = *dsp-- / st0;
		break;
	case 11: // <0
		st0 = 0 > st0;
		break;
	case 12: // exit
		pc = m[m[1]--];
		break;
	case 13: // echo
		putchar(st0);
		st0 = *dsp--;
		break;
	case 14: // key
		*++dsp = st0;
		st0 = getchar();
	case 15: // _pick
		st0 = dsp[-st0];
		break;
	}
}

int
main()
{
	at(3);
	at(4);
	at(1);
	w = *m;
	m[m[0]++] = 5;
	m[m[0]++] = 2;
	pc = *m;
	m[m[0]++] = w;
	m[m[0]++] = pc - 1;
	for (w = 6; w < 16;) {
		at(1);
		m[m[0]++] = w++;
	}
	m[1] = *m;
	for (*m += 512;; run(m[pc++]));

	return 0;
}
