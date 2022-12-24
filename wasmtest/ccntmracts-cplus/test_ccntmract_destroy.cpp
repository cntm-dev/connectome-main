#include<cntmiolib/cntmio.hpp>
using std::string;
using std::vector;

#define KEY_MIGRATE_STORE 0x23
#define VAL_MIGRAGE_STORE 0x249308

#define KEY_MIGRATE_STORE2 0x24
#define VAL_MIGRAGE_STORE2 0x372430

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
	void test_ccntmract_destroy(void) {
		int64_t b;
		int64_t a = VAL_MIGRAGE_STORE;
		key t = {KEY_MIGRATE_STORE};
		storage_put(t, a);
		check(storage_get(t, b), "get failed");
		check(b == a, "get wrcntm");

		a = VAL_MIGRAGE_STORE2;
		key t2 = {KEY_MIGRATE_STORE2};
		storage_put(t, a);
		check(storage_get(t, b), "get failed");
		check(b == a, "get wrcntm");

		cntmio::ccntmract_destroy();
		check(false, "should not be here");
	}

	string testcase(void) {
		return string(R"(
		[
			[{"method":"test_ccntmract_destroy", "param":"", "expected":""}
			]
		]
		)");
	}
};

cntmIO_DISPATCH( hello, (testcase)(test_ccntmract_destroy))
