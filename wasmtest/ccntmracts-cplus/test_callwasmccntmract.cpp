#include<cntmiolib/cntmio.hpp>
using std::string;
using std::vector;

namespace cntmio {
	struct test_conext {
		address admin;
		std::map<string, address> addrmap;
		cntmLIB_SERIALIZE( test_conext, (admin) (addrmap))
	};
};

using namespace cntmio;

class hello: public ccntmract {
	public:
	using ccntmract::ccntmract;

	int64_t call_wasm_ccntmract(int64_t a, int64_t b, test_conext &tc) {
		int64_t res;
		address test_add = tc.addrmap["test_add.wasm"];
		auto args = pack(string("add"), a, b);
		call_ccntmract(test_add, args, res);
		check(res == a + b, "call wasm ccntmract wrcntm");
		return res;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"needccntmext":true,"method":"call_wasm_ccntmract", "param":"int:1,int:2", "expected":"int:3"}
    	    ]
		]
		)");
	}

};

cntmIO_DISPATCH( hello,(testcase)(call_wasm_ccntmract))
