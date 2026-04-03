const SPECIALTIES = [
  'Cardiology', 'Dermatology', 'Neurology', 'Orthopedics',
  'Pediatrics', 'Psychiatry', 'Ophthalmology', 'Endocrinology',
  'Gastroenterology', 'Pulmonology', 'Acupuncture',
  'Veterinary Medicine', 'Physiotherapy', 'Nutrition',
];

const SEARCH_TERMS = [
  'heart', 'skin', 'brain', 'bones', 'children',
  'mental-health', 'eyes', 'diabetes', 'digestive', 'lungs',
  'Dr.', 'Silva', 'Santos', 'Oliveira', 'Souza',
  'cardiovascular', 'therapy', 'rehabilitation', 'diet',
  'clinical', 'diagnosis', 'treatment', 'holistic',
];

const KEYWORD_TERMS = [
  'heart', 'cardiovascular', 'hypertension', 'arrhythmia',
  'skin', 'dermatitis', 'acne', 'melanoma',
  'brain', 'nervous-system', 'epilepsy', 'migraine',
  'bones', 'joints', 'fractures', 'spine',
  'children', 'infant', 'vaccination', 'growth',
  'mental-health', 'anxiety', 'depression', 'therapy',
  'eyes', 'vision', 'glaucoma', 'cataract',
  'hormones', 'diabetes', 'thyroid', 'metabolism',
  'digestive', 'stomach', 'liver', 'intestine',
  'lungs', 'respiratory', 'asthma', 'pneumonia',
  'chinese-medicine', 'meridians', 'pain-relief', 'holistic',
  'animals', 'pets', 'surgery',
  'rehabilitation', 'movement', 'posture', 'recovery',
  'diet', 'weight', 'supplements',
];

const SORT_FIELDS = ['name', 'specialty', 'rating', 'created_at', 'updated_at'];
const SORT_ORDERS = ['asc', 'desc'];
const FILTER_FIELDS = ['specialty', 'keywords', 'name'];
const PAGE_SIZES = [5, 10, 20, 50];

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function searchTermOnly() {
  return {
    search_term: pick(SEARCH_TERMS),
    page_size: pick(PAGE_SIZES),
    sort: [{ field: pick(['rating', 'created_at', 'updated_at']), order: pick(SORT_ORDERS) }],
  };
}

function filterBySpecialty() {
  return {
    filters: [{ field: 'specialty', value: pick(SPECIALTIES) }],
    page_size: pick(PAGE_SIZES),
    sort: [{ field: pick(['rating', 'created_at', 'updated_at']), order: pick(SORT_ORDERS) }],
  };
}

function filterByKeyword() {
  return {
    filters: [{ field: 'keywords', value: pick(KEYWORD_TERMS) }],
    page_size: pick(PAGE_SIZES),
    sort: [{ field: pick(['rating', 'created_at', 'updated_at']), order: pick(SORT_ORDERS) }],
  };
}

function searchTermWithFilter() {
  return {
    search_term: pick(SEARCH_TERMS),
    filters: [{ field: 'specialty', value: pick(SPECIALTIES) }],
    page_size: pick(PAGE_SIZES),
    sort: [{ field: pick(['rating', 'created_at', 'updated_at']), order: pick(SORT_ORDERS) }],
  };
}

const STRATEGIES = [searchTermOnly, filterBySpecialty, filterByKeyword, searchTermWithFilter];

export function generateSearchPayload() {
  return pick(STRATEGIES)();
}
