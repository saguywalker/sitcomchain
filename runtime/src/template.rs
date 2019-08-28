/// A runtime module template with necessary imports

/// Feel free to remove or edit this file as needed.
/// If you change the name of this file, make sure to update its references in runtime/src/lib.rs
/// If you remove this file, you can remove those references


/// For more guidance on Substrate modules, see the example module
/// https://github.com/paritytech/substrate/blob/master/srml/example/src/lib.rs

use support::{decl_module, decl_storage, decl_event, ensure, StorageMap, StorageValue, dispatch::Result};
use system::ensure_signed;
use primitives::H256;
use runtime_primitives::traits::{As, BlakeTwo256, Hash};
use parity_codec::{Decode, Encode};

/// The module's configuration trait.
pub trait Trait: system::Trait {
	// TODO: Add other types and constants required configure this module.

	/// The overarching event type.
	type Event: From<Event<Self>> + Into<<Self as system::Trait>::Event>;
}

pub type PubKey = H256;
pub type CompetenceID = u32;	//start with 3 (eg. 3001)
pub type StudentID = u64;		//start with 4 (eg. 4123)
pub type ActivityID = u32;

#[cfg_attr(feature = "std", derive(Debug))]
#[derive(PartialEq, Eq, PartialOrd, Ord, Default, Clone, Encode, Decode, Hash)]
pub struct StaffAddCompetence<Hash>{
	pub id: Hash,
	pub student_id: StudentID,
	pub competence_id: CompetenceID,
	pub by: PubKey,
	pub semester: u16, // eg. semester 1 year 2019 => 12019
}

#[cfg_attr(feature = "std", derive(Debug))]
#[derive(PartialEq, Eq, PartialOrd, Ord, Default, Clone, Encode, Decode, Hash)]
pub struct AttendedActivity<Hash>{
	pub id: Hash,
	pub student_id: StudentID,
	pub activity_id: ActivityID,
	pub approver: PubKey,
}

#[cfg_attr(feature = "std", derive(Debug))]
#[derive(PartialEq, Eq, PartialOrd, Ord, Default, Clone, Encode, Decode, Hash)]
pub struct AutoAddCompetence<Hash>{
	pub id: Hash,
	pub student_id: StudentID,
	pub competence_id: CompetenceID,
	pub semester: u16,
}

/// This module's storage items.
decl_storage! {
	trait Store for Module<T: Trait> as SitcomStore {
		// map from Hash to struct
		StaffAddCompetenceMap get(staff_add_competence_map): map T::Hash => StaffAddCompetence<T::Hash>;
		AttendedActivityMap get(attended_activity_map): map T::Hash => AttendedActivity<T::Hash>;
		AutoAddCompetenceMap get(auto_add_competence_map): map T::Hash => AutoAddCompetence<T::Hash>;
		
		// map from student_id to competence_id and activity_id
		CollectedCompetencies get(competencies_from): map StudentID => Option<Vec<CompetenceID>>;		
		AttendedActivities get(activities_from): map StudentID => Option<Vec<ActivityID>>;

		// map each semester to each struct
		StaffAddCompetenciesSemester get(staff_add_competencies_semester): map u16 => Option<Vec<StaffAddCompetence<T::Hash>>>;
		AttendedActivitiesSemester get(attended_activities_semester): map u16 => Option<Vec<AttendedActivity<T::Hash>>>;
		AutoAddCompetenciesSemester get(auto_add_competencies_semester): map u16 => Option<Vec<AutoAddCompetence<T::Hash>>>;
	}
}

decl_module! {
	/// The module declaration.	
	pub struct Module<T: Trait> for enum Call where origin: T::Origin {
		// Initializing events
		// this is needed only if you are using events in your module
		fn deposit_event<T>() = default;

		// Just a dummy entry point.
		// function that can be called by the external world as an extrinsics call
		// takes a parameter of the type `AccountId`, stores it and emits an event
		pub fn do_something(origin, something: u32) -> Result {
			// TODO: You only need this if you want to check it was signed.
			let who = ensure_signed(origin)?;

			// TODO: Code to execute when something calls this.
			// For example: the following line stores the passed in u32 in the storage
			<Something<T>>::put(something);

			// here we are raising the Something event
			Self::deposit_event(RawEvent::SomethingStored(something, who));
			Ok(())
		}
	}
}

decl_event!(
	pub enum Event<T> where AccountId = <T as system::Trait>::AccountId {
		// Just a dummy event.
		// Event `Something` is declared with a parameter of the type `u32` and `AccountId`
		// To emit this event, we call the deposit funtion, from our runtime funtions
		SomethingStored(u32, AccountId),
	}
);

/// tests for this module
#[cfg(test)]
mod tests {
	use super::*;

	use runtime_io::with_externalities;
	use primitives::{H256, Blake2Hasher};
	use support::{impl_outer_origin, assert_ok};
	use runtime_primitives::{
		BuildStorage,
		traits::{BlakeTwo256, IdentityLookup},
		testing::{Digest, DigestItem, Header}
	};

	impl_outer_origin! {
		pub enum Origin for Test {}
	}

	// For testing the module, we construct most of a mock runtime. This means
	// first constructing a configuration type (`Test`) which `impl`s each of the
	// configuration traits of modules we want to use.
	#[derive(Clone, Eq, PartialEq)]
	pub struct Test;
	impl system::Trait for Test {
		type Origin = Origin;
		type Index = u64;
		type BlockNumber = u64;
		type Hash = H256;
		type Hashing = BlakeTwo256;
		type Digest = Digest;
		type AccountId = u64;
		type Lookup = IdentityLookup<Self::AccountId>;
		type Header = Header;
		type Event = ();
		type Log = DigestItem;
	}
	impl Trait for Test {
		type Event = ();
	}
	type TemplateModule = Module<Test>;

	// This function basically just builds a genesis storage key/value store according to
	// our desired mockup.
	fn new_test_ext() -> runtime_io::TestExternalities<Blake2Hasher> {
		system::GenesisConfig::<Test>::default().build_storage().unwrap().0.into()
	}

	#[test]
	fn it_works_for_default_value() {
		with_externalities(&mut new_test_ext(), || {
			// Just a dummy test for the dummy funtion `do_something`
			// calling the `do_something` function with a value 42
			assert_ok!(TemplateModule::do_something(Origin::signed(1), 42));
			// asserting that the stored value is equal to what we stored
			assert_eq!(TemplateModule::something(), Some(42));
		});
	}
}
