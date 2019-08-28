/// A runtime module template with necessary imports

/// Feel free to remove or edit this file as needed.
/// If you change the name of this file, make sure to update its references in runtime/src/lib.rs
/// If you remove this file, you can remove those references


/// For more guidance on Substrate modules, see the example module
/// https://github.com/paritytech/substrate/blob/master/srml/example/src/lib.rs

use support::{decl_module, decl_storage, decl_event, ensure, StorageMap, StorageValue, dispatch::Result, dispatch::Vec};
use system::ensure_signed;
//use primitives::H256;
use runtime_primitives::traits::{As, Hash};
use parity_codec::{Decode, Encode};

/// The module's configuration trait.
pub trait Trait: system::Trait {
	// TODO: Add other types and constants required configure this module.

	/// The overarching event type.
	type Event: From<Event<Self>> + Into<<Self as system::Trait>::Event>;
}

//pub type PubKey = H256;
pub type StudentID = u64;		
pub type CompetenceID = u16;	//start with 3 (eg. 30001)
pub type ActivityID = u32;		//start with 4 (4000000000 - 4999999999)

#[cfg_attr(feature = "std", derive(Debug))]
#[derive(PartialEq, Eq, PartialOrd, Ord, Default, Clone, Encode, Decode, Hash)]
pub struct StaffAddCompetence<Hash, AccountId>{
	pub id: Hash,
	pub student_id: StudentID,
	pub competence_id: CompetenceID,
	pub by: AccountId,
	pub semester: u16, // eg. semester 1 year 2019 => 12019
}

#[cfg_attr(feature = "std", derive(Debug))]
#[derive(PartialEq, Eq, PartialOrd, Ord, Default, Clone, Encode, Decode, Hash)]
pub struct AttendedActivity<Hash, AccountId>{
	pub id: Hash,
	pub student_id: StudentID,
	pub activity_id: ActivityID,
	pub approver: AccountId,
	pub semester: u16,
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
		StaffAddCompetenceMap get(staff_add_competence_map): map T::Hash => StaffAddCompetence<T::Hash, T::AccountId>;
		AttendedActivityMap get(attended_activity_map): map T::Hash => AttendedActivity<T::Hash, T::AccountId>;
		AutoAddCompetenceMap get(auto_add_competence_map): map T::Hash => AutoAddCompetence<T::Hash>;
		
		// map from student_id to competence_id and activity_id
		CollectedCompetencies get(competencies_from): map StudentID => Option<Vec<CompetenceID>>;		
		AttendedActivities get(activities_from): map StudentID => Option<Vec<ActivityID>>;

		// map each semester to each struct
		StaffAddCompetenciesSemester get(staff_add_competencies_semester): map u16 => Option<Vec<T::Hash>>;
		AttendedActivitiesSemester get(attended_activities_semester): map u16 => Option<Vec<T::Hash>>;
		AutoAddCompetenciesSemester get(auto_add_competencies_semester): map u16 => Option<Vec<T::Hash>>;
	}
}

decl_event!(
	pub enum Event<T>
	where
		<T as system::Trait>::AccountId
	{
		AddCompetece(StudentID, CompetenceID),
		AddCompetenceFromStaff(StudentID, CompetenceID, AccountId),
		ActivityApprove(StudentID, ActivityID, AccountId),
	}
);

decl_module! {
	/// The module declaration.	
	pub struct Module<T: Trait> for enum Call where origin: T::Origin {
		// Initializing events
		// this is needed only if you are using events in your module
		fn deposit_event<T>() = default;
		
		pub fn staff_add_competence(origin, student_id: StudentID, competence_id: CompetenceID, semester: u16, year: u16) -> Result{
			let by = ensure_signed(origin)?;
			ensure!(semester == 1 || semester == 2, "Semester should be only 1 or 2");
			ensure!(year >= 2000 && year <= 3000, "Improper academic year");
			let concat_semester = semester * 10000 + year;
			
			let random_hash = (<system::Module<T>>::random_seed(), &by, student_id, competence_id, concat_semester)
				.using_encoded(<T as system::Trait>::Hashing::hash);

			let new_struct = StaffAddCompetence{
				id: random_hash,
				student_id: student_id,
				competence_id: competence_id,
				by: by.clone(),
				semester: concat_semester,
			};

			let mut competencies_std: Vec<CompetenceID> = Self::competencies_from(student_id).unwrap_or(Vec::new());
			competencies_std.push(competence_id);

			let mut competencies_semester: Vec<T::Hash> = Self::staff_add_competencies_semester(concat_semester).unwrap_or(Vec::new());
			competencies_semester.push(random_hash);

			<StaffAddCompetenceMap<T>>::insert(random_hash, new_struct);

			<CollectedCompetencies<T>>::insert(student_id, competencies_std);
			<StaffAddCompetenciesSemester<T>>::insert(concat_semester, competencies_semester);

			Self::deposit_event(RawEvent::AddCompetenceFromStaff(student_id, competence_id, by));

			Ok(())
		}

		pub fn approve_activity(origin, student_id: StudentID, activity_id: ActivityID, semester: u16, year: u16) -> Result{
			let by = ensure_signed(origin)?;
			ensure!(semester == 1 || semester == 2, "Semester should be only 1 or 2");
			ensure!(year >= 2000 && year <= 3000, "Improper academic year");
			let concat_semester = semester * 10000 + year;
			
			let random_hash = (<system::Module<T>>::random_seed(), &by, student_id, activity_id, concat_semester)
				.using_encoded(<T as system::Trait>::Hashing::hash);

			let new_struct = AttendedActivity{
				id: random_hash,
				student_id: student_id,
				activity_id: activity_id,
				approver: by.clone(),
				semester: concat_semester,
			};

			let mut activities_std: Vec<ActivityID> = Self::activities_from(student_id).unwrap_or(Vec::new());
			activities_std.push(activity_id);

			let mut activities_semester: Vec<T::Hash> = Self::attended_activities_semester(concat_semester).unwrap_or(Vec::new());
			activities_semester.push(random_hash);

			<AttendedActivityMap<T>>::insert(random_hash, new_struct);

			<AttendedActivities<T>>::insert(student_id, activities_std);
			<AttendedActivitiesSemester<T>>::insert(concat_semester, activities_semester);

			Self::deposit_event(RawEvent::ActivityApprove(student_id, activity_id, by));

			Ok(())
		}

	}
}

impl<T: Trait> Module<T>{
	fn auto_add_competence(student_id: StudentID, competence_id: CompetenceID, semester: u16) -> Result{
		let random_hash = (<system::Module<T>>::random_seed(), student_id, competence_id, semester)
			.using_encoded(<T as system::Trait>::Hashing::hash);

		let new_struct = AutoAddCompetence{
			id: random_hash,
			student_id: student_id,
			competence_id: competence_id,
			semester: semester,
		};

		let mut competencies_std: Vec<CompetenceID> = Self::competencies_from(student_id).unwrap_or(Vec::new());
		competencies_std.push(competence_id);

		let mut competencies_semester: Vec<T::Hash> = Self::auto_add_competencies_semester(semester).unwrap_or(Vec::new());
		competencies_semester.push(random_hash);

		<AutoAddCompetenceMap<T>>::insert(random_hash, new_struct);
		
		<CollectedCompetencies<T>>::insert(student_id, competencies_std);
		<AutoAddCompetenciesSemester<T>>::insert(semester, competencies_semester);

		Self::deposit_event(RawEvent::AddCompetece(student_id, competence_id));

		Ok(())
	}
}

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
	type SitcomModule = Module<Test>;

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
			assert_ok!(SitcomModule::do_something(Origin::signed(1), 42));
			// asserting that the stored value is equal to what we stored
			assert_eq!(SitcomModule::something(), Some(42));
		});
	}
}
