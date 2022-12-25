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
	uint128_t test_native_cntm(string &method, address &from, address &to, asset &amount, test_conext &tc) {
		if (method == "balanceOf") {
			asset balance = cntm::balanceof(tc.admin);
			check(balance == 1000000000, "init balance wrcntm");
		} else if (method == "transfer") {
			/*keep admin alway initbalance.*/
			check(cntm::transfer(tc.admin, to, amount), "transfer failed");
			check(cntm::balanceof(to) == amount, "transfer amount wrcntm");
			check(cntm::transfer(to, tc.admin, amount), "transfer failed");
			check(cntm::balanceof(to) == 0, "transfer amount wrcntm");
		} else if (method == "approve") {
			/*keep admin alway initbalance.*/
			check(cntm::approve(tc.admin, from, amount),"approve failed");
			check(cntm::allowance(tc.admin, from) == amount, "allowance amount wrcntm");
			check(cntm::transferfrom(from, tc.admin, to, amount),"transferfrom failed");
			check(cntm::allowance(tc.admin, from) == 0, "allowance amount wrcntm");
			check(cntm::balanceof(to) == amount, "transfer amount wrcntm");
			check(cntm::transfer(to, tc.admin, amount), "transfer failed");
			check(cntm::balanceof(to) == 0, "transfer amount wrcntm");
			check(cntm::balanceof(from) == 0, "transfer amount wrcntm");
		}

		return 1;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"needccntmext":true, "method":"test_native_cntm", "param":"string:balanceOf,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"},
    	    {"env":{"witness":["Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT","Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV"]}, "needccntmext":true, "method":"test_native_cntm", "param":"string:transfer,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"},
    	    {"env":{"witness":["Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT","Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV"]}, "needccntmext":true, "method":"test_native_cntm", "param":"string:approve,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"}
    	    ]
		]
		)");
	}

};

cntmIO_DISPATCH( hello,(testcase)(test_native_cntm))
