#include<cntmiolib/cntmio.hpp>

using namespace cntmio;
using std::string;

class hello: public ccntmract {
	public:
	using ccntmract::ccntmract;
	int64_t add(int64_t a, int64_t b) {
		return a + b;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"env":{"witness":[]}, "method":"add", "param":"int:1, int:2", "expected":"int:3"}
    	    ]
		]
		)");
	}
};

cntmIO_DISPATCH( hello, (testcase)(add))
